# Internationalization (i18n) Example

This is an example app demonstrating translating rule sets for internationalization (i18n).

It uses the `text/messages` package to perform the translations and demonstrates:

1. **Standard Messages** - Built-in error messages from the error dictionary (e.g., `WithMinLen`, `WithMaxLen`)
2. **Custom Messages** - User-defined messages using `WithErrorMessage`

Both types of messages are translatable via i18n when you pass a `message.Printer` through the context using `rulecontext.WithPrinter`.

## How It Works

### Standard Messages
Standard validation errors (like min/max length) use messages defined in the error dictionary. These are automatically picked up by `gotext` for translation:

```go
var standardRuleSet = rules.String().
    WithMinLen(3).
    WithMaxLen(7)
```

### Custom Messages
You can define custom error messages using `WithErrorMessage`. These are also translatable:

```go
var customRuleSet = rules.String().
    WithMinLen(5).
    WithErrorMessage("too short", "username must be at least %d characters")
```

The long message (second argument) supports format specifiers like `%d` for dynamic values.

## Getting Started

1st, if you don't already, install `gotext`:
```bash
go install golang.org/x/text/cmd/gotext@latest
```

2nd, `cd` into the example directory and run `go generate` and `go run`:
```bash
# Create the catalog.go file and locales directory
go generate translations/translations.go
```

3rd, send the files in `locales` to a translation service then place the translated files in the directory.
For example purposes we'll just copy them as is and edit in place:
```bash
cp translations/locales/en-US/out.gotext.json translations/locales/en-US/messages.gotext.json
cp translations/locales/es-ES/out.gotext.json translations/locales/es-ES/messages.gotext.json
```

Edit the `messages.gotext.json` file as you see fit.

Every time you edit the files you will need to rerun `go generate`:
```bash
# Recreate the catalog.go file
go generate translations/translations.go
```

Don't worry, this will not overwrite your existing translations.

Now you can run the example:
```bash
# English (default)
go run app.go a ab abc abcd abcde

# Spanish
go run app.go -locale es-ES a ab abc abcd abcde
```

If the `text/messages` library can't find a translation it will default to the `en-US` (US English) locale.

## Tips for Production

- Customize the default (US English / `en-US`) error messages to fit your product voice.
- Update translations to support plural messages. For example: `1 characters` may become `1 character`.
- Use `WithErrorMessage` for domain-specific error messages that need custom wording.
