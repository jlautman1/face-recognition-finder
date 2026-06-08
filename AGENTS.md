# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project Scope

This repo has two entry points:
- Web app in [web/README.md](web/README.md): Go server + static frontend that calls the Python face-recognition script.
- CLI script in [README.md](README.md): direct Python usage for batch workflows.

## Start Here

- Read [README.md](README.md) for end-to-end setup and usage.
- Read [web/README.md](web/README.md) before changing web handlers, endpoints, or UI behavior.

## Environment And Run Commands

Python deps (repo root):

```bash
python3 -m venv venv
source venv/bin/activate
pip install deepface pillow tf-keras opencv-python
```

Web server:

```bash
cd web && go mod download
cd web && go run main.go
```

CLI:

```bash
python find_my_images.py <zip_file> <reference_path> <output_folder> [threshold]
```

## Validation Commands

Run these from repo root after starting the web server on `localhost:8080`:

```bash
node web/test_ui.js
node web/test-live-progress.spec.js
npx playwright test web/test-live-updates.spec.js
```

## Codebase Map

- Core Python processing: [find_my_images.py](find_my_images.py)
- Go orchestration and API: [web/main.go](web/main.go)
- Frontend assets: [web/static/index.html](web/static/index.html), [web/static/script.js](web/static/script.js), [web/static/style.css](web/static/style.css)
- Sample web tests: [web/test_ui.js](web/test_ui.js), [web/test-live-updates.spec.js](web/test-live-updates.spec.js)

## Repository-Specific Conventions

- Keep Python subprocess unbuffered in Go (`-u` flag) so progress updates remain real-time.
- Preserve thread-safe `JobStatus` updates via its mutex helper methods in [web/main.go](web/main.go).
- Frontend progress UX relies on periodic status polling; keep API response fields stable when editing status payloads.
- Python matching behavior uses Facenet512 + RetinaFace + cosine distance; do not silently change defaults without updating docs.

## Pitfalls To Avoid

- Large/generated data is intentionally not versioned; review [.gitignore](.gitignore) before adding files.
- `web/uploads/` and `web/outputs/` are runtime directories and can grow quickly.
- Many web tests depend on local sample assets under ignored folders (zip/reference photos). If missing, tests can fail even when code is correct.
- Initial model download can make first run slow.

## Change Guidance

- Prefer small, targeted edits.
- If changing API contracts or workflow, update both docs: [README.md](README.md) and [web/README.md](web/README.md).
- After meaningful behavior changes, run at least one relevant web test plus one manual flow check in browser.