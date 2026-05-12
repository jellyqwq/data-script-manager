package scheduler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func pythonRuntime(options RunOptions) string {
	if value := firstEnv(options, "DSM_PYTHON_BIN", "SCRIPT_PYTHON_BIN", "PYTHON_BIN"); value != "" {
		return value
	}

	if value := autoPythonRuntime(options); value != "" {
		return value
	}

	return "python3"
}

func autoPythonRuntime(options RunOptions) string {
	candidates := pythonCandidates()
	if len(candidates) == 0 {
		return ""
	}

	modules := pythonImportModules(options)
	if len(modules) == 0 {
		return candidates[0]
	}

	for _, candidate := range candidates {
		if pythonCanImport(candidate, modules) {
			return candidate
		}
	}
	return candidates[0]
}

func pythonCandidates() []string {
	seen := map[string]bool{}
	var candidates []string
	addPython := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			return
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			seen[path] = true
			candidates = append(candidates, path)
		}
	}
	addEnvPython := func(prefix string) {
		if prefix != "" {
			addPython(filepath.Join(prefix, "bin", "python"))
		}
	}

	addEnvPython(os.Getenv("VIRTUAL_ENV"))
	addEnvPython(os.Getenv("CONDA_PREFIX"))

	if cwd, err := os.Getwd(); err == nil {
		addPython(filepath.Join(cwd, ".venv", "bin", "python"))
		addPython(filepath.Join(cwd, "venv", "bin", "python"))
	}

	for _, dir := range condaEnvDirs() {
		addCondaEnvPythons(dir, addPython)
	}

	for _, name := range []string{"python", "python3"} {
		if path, err := exec.LookPath(name); err == nil {
			addPython(path)
		}
	}

	return candidates
}

func condaEnvDirs() []string {
	seen := map[string]bool{}
	var dirs []string
	addDir := func(dir string) {
		dir = strings.TrimSpace(dir)
		if dir == "" || seen[dir] {
			return
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}

	if condaExe := os.Getenv("CONDA_EXE"); condaExe != "" {
		addDir(filepath.Join(filepath.Dir(filepath.Dir(condaExe)), "envs"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		addDir(filepath.Join(home, "miniconda3", "envs"))
		addDir(filepath.Join(home, "anaconda3", "envs"))
	}
	addDir("/opt/miniconda3/envs")
	addDir("/opt/anaconda3/envs")
	addDir("/usr/local/miniconda3/envs")
	addDir("/usr/local/anaconda3/envs")

	return dirs
}

func addCondaEnvPythons(envDir string, addPython func(string)) {
	entries, err := os.ReadDir(envDir)
	if err != nil {
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		if entry.IsDir() {
			addPython(filepath.Join(envDir, entry.Name(), "bin", "python"))
		}
	}
}

func pythonCanImport(command string, modules []string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	code := strings.Join([]string{
		"import importlib, sys",
		"for mod in sys.argv[1:]:",
		"    importlib.import_module(mod)",
	}, "\n")

	args := append([]string{"-c", code}, modules...)
	cmd := exec.CommandContext(ctx, command, args...)
	return cmd.Run() == nil
}
