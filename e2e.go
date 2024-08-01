package endy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
)

type Tester struct {
	Tests  *[]Test
	Logger *zap.Logger
	Config *Config
}

type Config struct {
	Path      string
	Timeout   time.Duration
	BenchMode bool
}

type Header struct {
	Name      string `yaml:"name"`
	Value     string `yaml:"value"`
	EnvSecret string `yaml:"env_secret,omitempty"`
}

type Test struct {
	URL        string        `yaml:"url"`
	AssertCode int           `yaml:"assert_code"`
	Method     string        `yaml:"method"`
	Timeout    time.Duration `yaml:"timeout,omitempty"`
	Headers    []Header      `yaml:"headers,omitempty"`
	Body       string        `yaml:"body,omitempty"`
	Threads    string        `yaml:"threads,omitempty"`
	Duration   string        `yaml:"duration,omitempty"`
	Requests   string        `yaml:"requests,omitempty"`
}

func (t *Tester) Run() error {
	var (
		tests *[]Test

		to time.Duration
	)
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	yamlFile, err := os.ReadFile(pwd + string(os.PathSeparator) + t.Config.Path)
	if err != nil {
		return fmt.Errorf("read test configuration file: %v", err)
	}

	if err = yaml.Unmarshal(yamlFile, &tests); err != nil {
		return fmt.Errorf("unmarshall test configuration: %v", err)
	}

	preparedTests := prepareTests(tests)

	t.Tests = preparedTests

	if t.Config.Timeout != 0 {
		to = t.Config.Timeout
	} else {
		to = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	client := &http.Client{}

	if t.Config.BenchMode {
		color.Yellow("Running benchmark mode")
		for _, test := range *t.Tests {
			err = execBenchTest(t, test)
			if err != nil {
				t.Logger.Error("execute test", zap.String("url", test.URL), zap.Error(err))
				return err
			}
		}
		return nil
	}

	color.Green("Running API testing mode")
	for _, test := range *t.Tests {
		err = execAPITests(ctx, t, client, test)
		if err != nil {
			return err
		}
	}

	return nil
}

func New() *Tester {
	lg, err := zap.NewProduction(
		zap.WithCaller(false),
	)
	if err != nil {
		log.Fatal(err)
	}
	return &Tester{
		Logger: lg,
		Config: &Config{},
	}
}

func execAPITests(ctx context.Context, t *Tester, client *http.Client, test Test) error {
	var (
		buf bytes.Buffer
	)

	if &test.Body != nil {
		err := json.NewEncoder(&buf).Encode(test.Body)
		if err != nil {
			t.Logger.Error("json encode", zap.Error(err))
			return err
		}
	}

	req, err := http.NewRequestWithContext(ctx, test.Method, test.URL, &buf)
	if err != nil {
		t.Logger.Error("create http request", zap.Error(err))
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Logger.Error("send http request", zap.Error(err))
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			t.Logger.Error("close http body", zap.Error(err))
		}
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logger.Error("read response body", zap.Error(err))
		return err
	}
	bodyString := string(bodyBytes)

	switch {
	case resp.StatusCode == test.AssertCode:
		color.Green("Test passed")
		t.Logger.Info("Test passed", zap.String("url", test.URL), zap.String("method", test.Method), zap.String("body", buf.String()))
	default:
		color.Red("Test failed")
		t.Logger.Fatal("Test failed", zap.String("url", test.URL), zap.String("method", test.Method), zap.String("body", buf.String()), zap.Int("status_code", resp.StatusCode), zap.Error(err), zap.String("response_body", bodyString))
	}
	return nil
}

func execBenchTest(t *Tester, test Test) error {
	var (
		buf bytes.Buffer
	)

	if &test.Body != nil {
		err := json.NewEncoder(&buf).Encode(test.Body)
		if err != nil {
			t.Logger.Error("json encode", zap.Error(err))
			return err
		}
	}

	headers := prepareBenchHeaders(test.Headers)

	app := "bombardier"
	arg0 := "-c"
	arg1 := test.Threads
	arg2 := "-n"
	arg3 := test.Requests
	arg4 := "-m"
	arg5 := test.Method
	arg6 := "-b"
	arg7 := test.Body
	arg8 := "-k"
	arg9 := "-d"
	arg10 := test.Duration
	arg11 := "-H"
	arg12 := headers
	arg13 := "-p"
	arg14 := "r"
	arg15 := test.URL

	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10, arg11, arg12, arg13, arg14, arg15)

	t.Logger.Info("benchmarking", zap.String("url", test.URL), zap.String("command", cmd.String()))

	stdout, err := cmd.Output()
	if err != nil {
		t.Logger.Fatal("bench command failed", zap.Error(err), zap.String("output", string(stdout)))
	}

	t.Logger.Info("benchmark output", zap.String("output", string(stdout)))

	color.Green("benchmark passed")

	return nil
}

func (t *Tester) SetTimeout(timeout time.Duration) {
	t.Config.Timeout = timeout
}

func (t *Tester) SetConfigPath(path string) {
	t.Config.Path = path
}

func (t *Tester) SetBenchmarkMode() {
	t.Config.BenchMode = true
}

func prepareBenchHeaders(headers []Header) string {
	var headerString string
	for _, header := range headers {
		headerString += header.Name + ": " + header.Value + " "
	}
	return headerString
}

func prepareTests(tests *[]Test) *[]Test {
	for _, test := range *tests {
		prepareHeadersSecrets(test.Headers)
	}
	return tests
}

func prepareHeadersSecrets(headers []Header) []Header {
	for _, header := range headers {
		if header.EnvSecret != "" {
			value, found := os.LookupEnv(header.EnvSecret)
			if !found {
				log.Fatalf("Environment variable %s not found", header.EnvSecret)
			}

			header.Value = value
		}
	}
	return headers
}
