# Data Script Manager

Data Collection Script Management System Based on Fiber and Vue3 Framework.

## Docker

```bash
cp .env.example .env
docker compose up --build
```

The backend image contains a baseline Python runtime. For script-specific Python
packages, install them in the runtime yourself or build a custom backend image.
The platform auto-detects a Python interpreter from active virtualenv/Conda,
project `.venv`, common Conda env directories, and `PATH`. You can still force
one with `PYTHON_BIN` if needed.

Before a Python script runs, the backend checks the interpreter, compiles the
script, and imports top-level modules found in `import` / `from` statements. You
can override the import list with `DSM_PYTHON_IMPORTS=requests,Crypto` or skip
that check with `DSM_SKIP_IMPORT_CHECK=true`.

For Docker deployments, a script owner can create a persistent virtualenv under
the mounted storage directory and point the script env group to it:

```bash
docker compose exec backend python -m venv /app/storage/venvs/ikuuu
docker compose exec backend /app/storage/venvs/ikuuu/bin/pip install requests pycryptodome
```

Then add this to the script env group:

```text
PYTHON_BIN=/app/storage/venvs/ikuuu/bin/python
```
