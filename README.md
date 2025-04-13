
# ez-scp 🚀

**ez-scp** is a lightweight, interactive terminal-based SCP (Secure Copy) client built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). It supports file and folder upload/download over SSH using an intuitive TUI and offers smart autocomplete using persistent history.

---

## ✨ Features

- 🔐 SCP over SSH using username/password
- 📁 Upload or download files and zipped folders
- 💾 Persistent input history (auto-suggestions)
- 🧠 Smart autocomplete with `Tab` or `Right Arrow`
- 🖥️ Minimal terminal UI using Bubble Tea
- ⛓️ No external dependencies needed at runtime

---

## 🛠️ Build

You need [Go 1.20+](https://go.dev/dl/) installed.

```bash
git clone https://github.com/yourname/ez-scp.git
cd ez-scp
go build -o ez-scp main.go
```

> 🔒 The resulting `ez-scp` binary can be distributed and used on other machines without requiring Go.

---

## 🚀 Run

```bash
./ez-scp
```

---

## 🖥️ UI Guide

You’ll see 7 input fields:

1. **FilePath** – Path to the local file/folder
2. **TargetPath** – Remote path (target on server)
3. **FileMode** – `0` for file, `1` for folder
4. **Mode** – `0` for upload, `1` for download
5. **Username** – SSH username
6. **Address** – SSH IP/domain (without port)
7. **Password** – SSH password (hidden)

Navigate with:
- `Tab` / `Right Arrow` for autocomplete
- `Up/Down` arrows to move between fields
- `Enter` to submit

---

## 📦 Examples

### ✅ Upload a file
- FilePath: `./myfile.txt`
- TargetPath: `/home/user`
- FileMode: `0`
- Mode: `0`
- Username: `root`
- Address: `192.168.1.1`
- Password: `yourpassword`

➡️ Uploads `myfile.txt` to `/home/user` on the server.

---

### ✅ Upload a folder
- FilePath: `./project`
- TargetPath: `/home/user`
- FileMode: `1`
- Mode: `0`
- Username: `admin`
- Address: `example.com`
- Password: `admin123`

➡️ Compresses `project/` into `project.zip` and uploads it.

---

### ✅ Download a file
- FilePath: `./output.txt`
- TargetPath: `/var/data`
- FileMode: `0`
- Mode: `1`
- Username: `user`
- Address: `10.0.0.5`
- Password: `pass123`

➡️ Downloads `/var/data/output.txt` from server to current folder.

---

## 🧠 Autocomplete History

Your input history is saved in `history.json`. This includes:
- Paths
- Target folders
- SSH usernames
- Addresses

Use `Tab` or `→` to quickly fill matching values.

---

## 📂 Folder Compression

When FileMode is set to `1`, the folder is automatically zipped before transfer.

---

## ⚠️ Limitations

- Only password-based auth (no key auth)
- Download only works for single files (not folders)
- No resume on interruption
- No folder downloading. will do it later

---

## 🧾 License

MIT © Muzzii255

---

## 🤝 Contributions

PRs are welcome! Feel free to fork and improve the tool.