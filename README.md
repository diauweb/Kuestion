# Kuestion

## Setup
Fill these environment variables in your `.env` file.
```sh
PORT=25566
# Your GitHub Personal Token to file issue
GH_PAT=0123456789abcdef000000000000000000000000
# Target GitHub Repository
GH_REPO=user/repo
# hCaptcha Configurations
HCAPTCHA_SITE_KEY=
HCAPTCHA_SECRET_KEY=
# Callback url to redirect when ok, with a ?ok=1 query
SUCCESS_CALLBACK=/
# When enabled, host files in /static.
STANDALONE=false
# Secret to trigger webhook
WEBHOOK_SECRET=1145141919810893
# Render **Answers** as Markdown
# Questions are always left as-is due to potential xss issues
BOX_RENDER_MD=true
```

## Routes
- `/sbmt` post to submit box
- `/box` render `tmpl/bako.gohtml`, show issues labeled with `publish`
- `/trigger` webhook endpoint
