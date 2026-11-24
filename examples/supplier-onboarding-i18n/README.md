# Supplier onboarding in 15 locales with TeaQL Go

The [KSML model](model.xml) describes a warehouse with suppliers. The [semantic model](semantics.json) sends two registrations through the generated workspace. One lacks a supplier name. The other submits 3 as the minimum order quantity, below the allowed value of 10. The application attempts audited saves and uses generated Q to confirm that SQLite stores neither invalid supplier.

## Run and select a locale

Go 1.18 or newer, Python 3.9 or newer and a C compiler for the SQLite driver. The verified compiler was Go 1.18.1. This example uses local Go runtime 0.6.0 source. The launcher creates a disposable `go.work`, links the generated sibling library, and resolves the runtime module to the enclosing source checkout.

```sh
bash examples/supplier-onboarding-i18n/run.sh
bash examples/supplier-onboarding-i18n/run.sh --locale fr
bash examples/supplier-onboarding-i18n/run.sh --all
bash examples/supplier-onboarding-i18n/run.sh --edge-cases
bash examples/supplier-onboarding-i18n/run.sh --verify
```

The commands run from the runtime repository root. The output is UTF-8 JSON Lines. Each validation row gives the effective locale, stable message key, canonical field, native rule, arguments and localized message. `--all` uses the 15 locales from the semantic model in one process. The final command runs the full set twice against one SQLite file and runs the locale boundary checks.

The app selects the locale with `context.SetLocaleCode(code)` and reads `context.Language()`. A new context defaults to `en`. The parser trims, normalizes underscores and case, and resolves known aliases such as `ZH_hant` to `zh-TW`. An empty or unsupported code raises `*runtime.UnsupportedLocaleError` and preserves the previous locale. Single-language mode prints the rejection and exits 2. Locale selection comes from the trusted CLI and context. Environment variables and business input do not set it.

`GO` defaults to `go` and `TEAQL_GO_EXAMPLE_SOURCE_ROOT` defaults to the enclosing repository.

## How the message is produced

1. The console builds its `UserContext`, installs generated model metadata, and explicitly initializes SQLite. It sets the locale before the generated business call.
2. The real model check rejects the invalid registration. The generated save returns a `*runtime.RuntimeError` of type `Check`. The original `CheckResults` report the canonical field, native rule and system bound. The app normalizes the latter to the number 10 in shared output.
3. The renderer chooses the locale and a template for the field's location shape. It binds the rendered field, model path and applicable checker parameters, including the lower bound.
4. The sample attaches the generated application catalog through `context.InstallI18nCatalog(catalog) after runtime.ParseI18nCatalog(data, runtime.BuiltinI18nCatalog())`. Its Vietnamese message entries are absent. Built-in Vietnamese supplies the same two messages after attachment. For this natural checker message, selection tries the application's shape-specific override, the application legacy template, and the built-in natural template at the active locale, then repeats that order for English. The generic invalid rule can supply the final natural template. Direct lookup through the legacy message catalog selects application and built-in entries at the current locale first, then application and built-in English, then the stable key.

The independent French context remains French through the main context's Japanese, Arabic and Japanese sequence. Domain labels without an available translation can remain in English. The generated provider reads `SUPPLIER_ONBOARDING_SERVICE_CORE_DATABASE_URL`, which this app supplies. The observed Go messages add a localized current-object location line. The numeric checker bound is a string in the native result and a number in the shared argument. [Locale source](../../runtime/i18n.go) · [Message lookup source](../../runtime/i18n.go) · [Natural templates source](../../runtime/checker_message_catalog.go)

## Run the same scene in another runtime

Keep the same KSML and semantic inputs when switching to its sibling checkout. Each runtime uses its own setup and call style, while the error keys, field paths and minimum bound stay stable.

[Java](../../../java/examples/supplier-onboarding-i18n/README.md) · [Rust](../../../rust/examples/supplier-onboarding-i18n/README.md) · [TypeScript](../../../typescript/examples/supplier-onboarding-i18n/README.md) · [Go](./README.md) · [Python](../../../python/examples/supplier-onboarding-i18n/README.md) · [C#/.NET](../../../dotnet/examples/supplier-onboarding-i18n/README.md) · [Swift](../../../swift/Examples/SupplierOnboardingI18n/README.md)

## Add a locale

Define `LocaleItalian Locale = "it"`, append it to `Locales` in `runtime/i18n.go`, and register any regional aliases. Update the built-in resources `builtin-messages-v1.json`, `natural-messages/builtin-v1.json` and `natural-messages/presentation-defaults-v1.json` under `runtime`. Supply natural forms for the supported shapes and add presentation words, such as the name for the current object and the measurement units. For a legacy required-message entry the identifier and placeholder stay the same.

```json
{
  "it": {
    "messages": {"checker.required": "Il campo {location} è obbligatorio"},
    "vocabulary": {}
  }
}
```

Keep placeholders such as `{field}`, `{location}` and `{system}` aligned with their respective catalog schemas. Add the locale to the generator and shared semantic model, regenerate the workspaces, then run the same save, alias, fallback and isolation checks. To customize text for an already supported locale, attach an explicit application catalog.
