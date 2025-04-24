[![Build](https://github.com/mbevc1/aws-ssm-util/actions/workflows/build.yaml/badge.svg)](https://github.com/mbevc1/aws-ssm-util/actions/workflows/build.yaml)

# aws-ssm-util

A lightweight, user-friendly CLI tool for managing AWS Systems Manager (SSM) Parameter Store using YAML config files. It supports uploading, downloading, deleting, tree visualization, and secure secret handling with smart heuristics.

---

## 🚀 Features

- ✅ Upload YAML configs to SSM Parameter Store
- 📥 Download SSM parameters into a YAML file
- 🔐 Upload secrets as SecureStrings (manual or smart detection)
- 🌲 Visualize parameters in a tree structure
- 🔄 Round-trip safe: YAML to SSM and back
- 🗑️ Delete parameters based on YAML keys
- 🎨 Colored CLI output with SecureString locks (🔒)
- ⚙️ Bash and Zsh autocompletions
- � version support with `--version` or `-v`

---

## 💶 Installation

1. Download from the [releases](https://github.com/mbevc1/aws-ssm-util/releases)
2. Run `aws-ssm-util -v` to check if it's working correctly.
3. Enjoy!

or build manually using:

```bash
git clone https://github.com/mbevc1/aws-ssm-util.git
cd aws-ssm-util
make build
```

---

## 🧪 Example Usage

### Load
```bash
aws-ssm-util load -f config.yaml -p /myapp
```

### Load with all as secrets
```bash
aws-ssm-util load -f secrets.yaml -p /secure --secure
```

### Load with smart secret detection
```bash
aws-ssm-util load -f config.yaml -p /myapp --smart-secure
```

### Save
```bash
aws-ssm-util save -p /myapp -o downloaded.yaml
```

### Delete
```bash
aws-ssm-util delete -f config.yaml -p /myapp
```

### Tree from SSM
```bash
aws-ssm-util tree -p /myapp
```

### Tree from YAML
```bash
aws-ssm-util yaml-tree -f config.yaml
```

#### Example output

```yaml
root
└── api
    ├── endpoint = https://api.example.com
    ├── token 🔒 = abc123xyz
└── app_name = my-service
└── db
    ├── host = localhost
    ├── password 🔒 = supersecret
    ├── port = 5432
    ├── user = admin
└── debug = true
└── servers
    ├── 0 = web-1.local
    ├── 1 = web-2.local
└── timeout_seconds = 2.5
```

---

## 🔐 SecureString Support

- Use `--secure` / `-s` to upload all values as SecureStrings
- Use `--auto-secure` / `-a` to auto-detect secrets based on key names (e.g., `password`, `secret`, `token`, etc.)
- Secure parameters are shown with a 🔒 lock in `load`, `tree`, `save`, and `delete`

---

## 🧩 Bash & Zsh Completion

### Bash
```bash
source <(aws-ssm-util completion bash)
# Or persist:
aws-ssm-util completion bash > /etc/bash_completion.d/aws-ssm-util
```

### Zsh
```bash
echo "autoload -U compinit; compinit" >> ~/.zshrc
aws-ssm-util completion zsh > ${fpath[1]}/_aws-ssm-util
```

---

## 🧰 Example YAML File

```yaml
debug: true
db:
  host: localhost
  password: mysecret
servers:
  - web1
  - web2
```

---

## Contributing

Report issues/questions/feature requests on in the [issues](https://github.com/mbevc1/aws-ssm-util/issues/new) section.

Full contributing [guidelines are covered here](.github/CONTRIBUTING.md).

## Authors

* [Marko Bevc](https://github.com/mbevc1)
* Full [contributors list](https://github.com/mbevc1/aws-ssm-util/graphs/contributors)

## 📄 License

MPL-2.0 Licensed. See [LICENSE](LICENSE) for full details.
<!-- https://choosealicense.com/licenses/ -->
