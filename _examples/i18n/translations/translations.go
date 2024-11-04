// This package provides translations for the i18n example code.
// See README.md for details.
package translations

// If you have more than one entry point you can include them both here.
// The default language does not have to be included in the output but we include it here so that you can customize the en-US translations.
//go:generate gotext -srclang=en-US update -out=catalog.go -lang=en-US,es-ES proto.zip/studio/validate/_examples/i18n
