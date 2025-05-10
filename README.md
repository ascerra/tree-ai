# tree-ai 🧠🌲

`tree-ai` is a command-line tool that enhances the `tree` output with AI-generated descriptions of each file and folder using IBM's Granite 4.0 Tiny Preview model. It works entirely offline after install, and outputs AI-enriched directory listings.

---

## 🚀 Install and Run

### 🔁 Option 1: One-liner install and run (via `curl`)

```bash
curl -sSf https://raw.githubusercontent.com/ascerra/tree-ai/main/install-and-run.sh | bash -s ./my-project
```

This will:
- Clone the repo into a temp directory
- Set up a local Python virtual environment
- Download the Granite model via Hugging Face
- Build the Go CLI and AI runner
- Run `tree-ai` on the specified directory

---


### 🛠 Option 2: Manual install for development

```bash
# Step 1: Clone the repo
git clone https://github.com/ascerra/tree-ai.git

# Step 2: Move into the project directory
cd tree-ai

# Step 3: Build and install dependencies
make install

# Step 4: Activate your Python virtual environment
source .venv/bin/activate

# Step 5: Run tree-ai on a directory
./bin/tree-ai ./my-project
```

---

## 💡 What You Get

```bash
./
├── cmd (CLI entrypoint)
├── internal/
│   ├── ai/ (AI integration using IBM Granite)
│   └── tree/ (tree display logic)
├── model/
│   ├── granite-runner.go (Go CLI wrapper around Granite model)
│   └── granite_infer.py (Python script running IBM Granite inference)
```

When you run `tree-ai`, you’ll see output like:

```
.
├── main.go (Handles CLI setup and command execution)
├── internal (Contains logic for AI and tree display)
│   └── ai (Integrates IBM Granite model for description generation)
```

---

## 📦 Project Structure

- `main.go` – entrypoint, calls `cmd.Execute()`
- `cmd/` – Cobra CLI setup
- `internal/` – Go logic for tree rendering and AI prompt integration
- `model/` – AI runner files (`granite-runner.go` and `granite_infer.py`)
- `.venv/` – Python virtualenv created by `make install`

---

## 🛠 Requirements

- Go 1.20+
- Python 3.8+
- Internet access for first-time model download via Hugging Face

---

## 🔐 Offline by Default

Once installed, all AI inference runs locally using the IBM Granite 4.0 Tiny Preview model. No data is sent to any server after setup.

---

## 🧹 Optional Cleanup

```bash
make clean
rm -rf .venv/
```

---

## 🧪 Testing

```bash
make test
make cover  # to see test coverage
```

---

## ✨ Coming Soon

- Shell completion
- Model choice via flags (e.g. --model)
- Built-in file filters and summarization toggles