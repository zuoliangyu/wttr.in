# Translations

This directory contains **all translation data** for wttr.in.

## Directory Structure

```bash
share/translations/
├── en/                          # Language directory
│   ├── conditions.txt           # Weather condition codes + names
│   ├── messages.json            # General UI messages and captions
│   ├── v1.json                  # v1-specific strings (time of day, etc.)
│   ├── v2.json                  # v2-specific strings
│   └── metadata.json            # Language metadata + translators
├── fr/
│   ├── conditions.txt
│   ├── messages.json
│   ├── v1.json
│   ├── v2.json
│   └── metadata.json
└── ...
```

### File Purposes

| File                | Purpose |
|---------------------|-------|
| `conditions.txt`    | Weather condition translations (old format, one line per code) |
| `messages.json`     | All general messages, captions, help texts, error messages |
| `v1.json`           | Strings used only by the legacy v1 renderer |
| `v2.json`           | Strings used by the current v2 renderer |
| `metadata.json`     | Language name, locale, translators, full/partial status |

## How to Contribute

### Adding or Improving a Translation (Recommended way)

The easiest and cleanest way is to **copy an existing well-translated language** and adapt it.

#### Example: Improving French or adding a new language

1. **Copy the French directory** (best starting point for most languages):
   ```bash
   cp -r share/translations/fr share/translations/XX
   ```
   (replace `XX` with the 2-letter code of your language, e.g. `gu`, `hi`, `pl`, `tr`, etc.)

2. **Rename and edit the files inside the new folder**:
   - Go into `share/translations/XX/`
   - Edit `metadata.json`:
     ```json
     {
       "code": "xx",
       "name": "Your Language Name",
       "locale": "xx_XX",
       ...
     }
     ```

   - Edit `messages.json`, `v1.json`, `v2.json` and translate the values.

   - (Strongly recommended) Also edit `conditions.txt` — translate everything after the second `:` on each line.

3. **Test your changes**: (to be fixed)
   ```bash
   curl "127.0.0.1/Paris?lang=xx"
   ```

4. **Commit and push**:
   ```bash
   git add share/translations/xx
   git commit -m "Add/improve xx translation"
   git push
   ```

5. Open a **Pull Request** on GitHub.

### Quick Checklist Before Submitting

- `metadata.json` contains correct `name`, `locale`, and translators
- All strings in `messages.json`, `v1.json`, `v2.json` are translated
- `conditions.txt` is updated (very visible to users)

---

**We welcome partial translations!**  
Even if only 30–40% is done, it’s still very useful.
