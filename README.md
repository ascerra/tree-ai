# tree-ai

**tree-ai** is a command-line tool that augments the traditional `tree` command with **AI-generated descriptions** for files and folders. It uses IBM Granite language models to provide concise summaries of each element in your directory tree, helping you quickly understand unfamiliar projects.

## Features

- 🧠 AI-generated summaries of files and directories
- 🔍 Fully recursive traversal with `--max-depth`
- 📝 Customizable prompt instructions `--prompt`
- ✂️ Can strictly enforce one line with `--truncate`
- 🧰 Support for both local and remote AI models
- 📦 Works offline if model is cached

## Installation

### 🔧 For Developers (from source)
Clone the repository and install from source:

```bash
git clone https://github.com/your-org/tree-ai.git
cd tree-ai
make install
source .venv/bin/activate
```

## Usage

To use your own model, you must export your API key and provide both `--model` and `--endpoint`:

```bash
export TREE_AI_API_KEY=<your-api-key>
bin/tree-ai ./ \
  --endpoint=https://your-model-endpoint.example.com/v1/completions \
  --model=your-model-id
```

Show up levels based on preference:

```bash
bin/tree-ai ./ --max-depth=3
```

Use a custom instruction for summarization with `--prompt`:

```bash
bin/tree-ai ./ --prompt "Summarize what this file contributes to the project."
```

Use `--truncate` to keep summaries to one line (useful for compact output):

```bash
bin/tree-ai ./ --truncate
```


Include hidden files and directories (like `tree -a`):

```bash
bin/tree-ai ./ --include-dotfiles
```

Enable verbose output for debugging:

```bash
bin/tree-ai ./ --verbose
```

## Output Example

```bash
❯ bin/tree-ai ./ --endpoint="<model endpoint>" --truncate          
⚠️  AI-generated summaries may be inaccurate or outdated. Always verify important details.
└── LICENSE ➤ grants users permission to use, modify, and distribute the project's software
└── Makefile ➤ as a build and testing automation tool for the tree-ai project
└── README.md ➤ This file serves as the project's documentation and user guide
├── bin ➤ serves as a centralized location for executable scripts and utilities
│   └── tree-ai ➤ outlines the architecture and details for integrating an AI model
├── cmd ➤ houses the command-line interface (CLI) implementation for the project
│   └── root.go ➤ The purpose of this file is to define the command-line interface
```

## Testing

```bash
make test
```

## License

MIT License
