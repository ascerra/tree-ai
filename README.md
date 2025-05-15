# tree-ai

**tree-ai** is a command-line tool that augments the traditional `tree` command with **AI-generated descriptions** for files and folders. It uses IBM Granite language models to provide concise summaries of each element in your directory tree, helping you quickly understand unfamiliar projects.

## Features

- 🧠 AI-generated summaries of files and directories
- 🔍 Fully recursive traversal with `--max-depth`
- 🧰 Support for both local and remote IBM Granite models
- 📦 Works offline if model is cached
- 📝 Customizable prompt instructions

## Installation

### 🚀 For Users (via curl)

You can install the latest release directly:

```bash
curl -sSL https://raw.githubusercontent.com/ascerra/tree-ai/main/install-and-run.sh | bash -s 
```

This will download the latest binary to your local `bin/` folder and make it executable.

To install dependencies for the local Python-based runner:

```bash
make deps
```

### 🔧 For Developers (from source)

```bash
git clone https://github.com/your-org/tree-ai.git
cd tree-ai
make install
```


## Usage

```bash
bin/tree-ai ./
```

Show up to 3 levels deep:

```bash
bin/tree-ai ./ --max-depth=3
```

Include hidden files and directories (like `tree -a`):

```bash
bin/tree-ai ./ --include-dotfiles
```

Use a specific remote model and endpoint:

```bash
bin/tree-ai ../incubator-devlake \
  --endpoint=https://granite-8b-code-instruct-maas-apicast-production.apps.prod.rhoai.rh-aiservices-bu.com:443/v1/completions \
  --model=granite-8b-code-instruct-128k
```

Use a custom instruction for summarization:

```bash
bin/tree-ai ./ --prompt-instruction "Summarize what this file contributes to the project."
```

Enable verbose output for debugging:

```bash
bin/tree-ai ./ --verbose
```

## Output Example

```bash
~/development/AI/project-tree-ai/tree-ai ❯ bin/tree-ai ./ --max-depth=2
└── 📄 LICENSE      This file, "LICENSE", is a legal notice that grants permission to use, modify, distribute, and sublicense the project's software, adhering to the MIT License terms, while limiting liability for any claims or damages.
└── 📄 Makefile     This Makefile outlines the build, testing, and installation processes for the "tree-ai" project, including its main Go binary, Python dependencies, and the "Granite" model runner.
└── 📄 README.md    The purpose of this file is to provide comprehensive documentation for installing, running, and understanding the structure and functionality of the `tree-ai` project.
└── 💼 bin          This directory contains compiled binaries used for local execution and testing.
└── 💼 cmd          This directory contains the Cobra-based CLI entrypoint logic.
└── 💼 internal     Internal Go packages for AI integration and tree traversal logic.
└── 📄 main.go      Main entry point for the tree-ai command-line interface.
```

## Testing

```bash
make test
```

## License

MIT License
