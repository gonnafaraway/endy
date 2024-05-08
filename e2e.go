package endy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Tester struct {
	Tests  *[]Test
	Logger *zap.Logger
	Config *Config
}

type Config struct {
	Path    string
	Timeout time.Duration
}

type Header struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Test struct {
	URL        string        `yaml:"url"`
	AssertCode int           `yaml:"assert_code"`
	Method     string        `yaml:"method"`
	Timeout    time.Duration `yaml:"timeout,omitempty"`
	Headers    []Header      `yaml:"headers,omitempty"`
	Body       string        `yaml:"body,omitempty"`
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

func (t *Tester) SetTimeout(timeout time.Duration) {
	t.Config.Timeout = timeout
}

func (t *Tester) SetConfigPath(path string) {
	t.Config.Path = path
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
	t.Tests = tests

	if t.Config.Timeout != 0 {
		to = t.Config.Timeout
	} else {
		to = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	client := &http.Client{}

	for _, test := range *t.Tests {
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
	}
	return nil
}
