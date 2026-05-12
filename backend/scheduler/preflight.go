package scheduler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func preflightScript(runID primitive.ObjectID, options RunOptions, command string, stateDir string) error {
	if err := checkRuntimeBinary(command); err != nil {
		return err
	}

	if strings.ToLower(filepath.Ext(options.ScriptPath)) != ".py" {
		return nil
	}

	env := buildProcessEnv(options, runID, stateDir)
	env = appendPythonPath(env, filepath.Dir(options.ScriptPath))

	if err := runPreflightCommand(stateDir, env, command, "-m", "py_compile", options.ScriptPath); err != nil {
		return fmt.Errorf("Python 语法检查失败: %w", err)
	}

	if truthyEnv(options, "DSM_SKIP_IMPORT_CHECK", "SCRIPT_SKIP_IMPORT_CHECK") {
		return nil
	}

	modules := pythonImportModules(options)
	if len(modules) == 0 {
		return nil
	}
	if err := runPythonImportCheck(stateDir, env, command, modules); err != nil {
		return err
	}
	return nil
}

func checkRuntimeBinary(command string) error {
	if filepath.IsAbs(command) || strings.Contains(command, string(os.PathSeparator)) {
		info, err := os.Stat(command)
		if err != nil {
			return fmt.Errorf("运行时不存在: %s", command)
		}
		if info.IsDir() {
			return fmt.Errorf("运行时不是可执行文件: %s", command)
		}
		if info.Mode()&0111 == 0 {
			return fmt.Errorf("运行时没有执行权限: %s", command)
		}
		return nil
	}

	if _, err := exec.LookPath(command); err != nil {
		return fmt.Errorf("找不到运行时命令: %s", command)
	}
	return nil
}

func runPreflightCommand(dir string, env []string, command string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("检查超时: %s %s", command, strings.Join(args, " "))
	}
	if err != nil {
		text := strings.TrimSpace(string(output))
		if len(text) > 1200 {
			text = text[len(text)-1200:]
		}
		if text == "" {
			text = err.Error()
		}
		return fmt.Errorf("%s", text)
	}
	return nil
}

func runPythonImportCheck(dir string, env []string, command string, modules []string) error {
	code := strings.Join([]string{
		"import importlib, sys",
		"missing = []",
		"for mod in sys.argv[1:]:",
		"    try:",
		"        importlib.import_module(mod)",
		"    except Exception as exc:",
		"        missing.append(f'{mod}: {exc.__class__.__name__}: {exc}')",
		"if missing:",
		"    print('缺少或无法导入 Python 模块:\\n' + '\\n'.join(missing))",
		"    sys.exit(1)",
	}, "\n")

	args := append([]string{"-c", code}, modules...)
	if err := runPreflightCommand(dir, env, command, args...); err != nil {
		return fmt.Errorf("Python 依赖检查失败: %w", err)
	}
	return nil
}

func pythonImportModules(options RunOptions) []string {
	spec := firstEnv(options, "DSM_PYTHON_IMPORTS", "SCRIPT_PYTHON_IMPORTS", "PYTHON_IMPORTS")
	if spec != "" {
		if isDisabledSpec(spec) {
			return nil
		}
		return splitModuleSpec(spec)
	}

	modules, err := detectPythonImports(options.ScriptPath)
	if err != nil {
		return nil
	}
	return modules
}

func detectPythonImports(scriptPath string) ([]string, error) {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var modules []string
	for _, line := range strings.Split(string(content), "\n") {
		line = stripInlineComment(line)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "import "):
			for _, item := range strings.Split(strings.TrimPrefix(line, "import "), ",") {
				module := strings.TrimSpace(strings.Split(item, " as ")[0])
				if appendModule(seen, module) {
					modules = append(modules, module)
				}
			}
		case strings.HasPrefix(line, "from "):
			parts := strings.Fields(line)
			if len(parts) >= 4 && parts[0] == "from" && parts[2] == "import" && !strings.HasPrefix(parts[1], ".") {
				if appendModule(seen, parts[1]) {
					modules = append(modules, parts[1])
				}
			}
		}
	}
	return modules, nil
}

func stripInlineComment(line string) string {
	if index := strings.Index(line, "#"); index >= 0 {
		return line[:index]
	}
	return line
}

func appendModule(seen map[string]bool, module string) bool {
	module = strings.TrimSpace(module)
	if !isPythonModuleName(module) || seen[module] {
		return false
	}
	seen[module] = true
	return true
}

func splitModuleSpec(spec string) []string {
	parts := strings.FieldsFunc(spec, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == ' '
	})
	seen := map[string]bool{}
	var modules []string
	for _, part := range parts {
		if appendModule(seen, part) {
			modules = append(modules, part)
		}
	}
	return modules
}

func isPythonModuleName(module string) bool {
	if module == "" {
		return false
	}
	for _, part := range strings.Split(module, ".") {
		if part == "" {
			return false
		}
		for index, r := range part {
			valid := r == '_' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || index > 0 && r >= '0' && r <= '9'
			if !valid {
				return false
			}
		}
	}
	return true
}

func appendPythonPath(env []string, scriptDir string) []string {
	for _, item := range env {
		key, value, ok := strings.Cut(item, "=")
		if ok && key == "PYTHONPATH" {
			return append(env, "PYTHONPATH="+scriptDir+string(os.PathListSeparator)+value)
		}
	}
	return append(env, "PYTHONPATH="+scriptDir)
}

func firstEnv(options RunOptions, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(options.EnvVars[name]); value != "" {
			return value
		}
	}
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func truthyEnv(options RunOptions, names ...string) bool {
	value := strings.ToLower(firstEnv(options, names...))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func isDisabledSpec(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "0" || value == "false" || value == "none" || value == "off" || value == "skip"
}
