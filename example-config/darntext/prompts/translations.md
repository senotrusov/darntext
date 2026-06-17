## Project-specific guidelines

* Ensure that all changes to the web interface fully internationalize every
  user-facing string introduced or modified by the change. This includes UI
  elements, alt text, accessibility labels, and error messages originating
  from both the business logic and the storage layer. Wrap all such strings
  with 'gettext' and use English directly in the source code as the 'msgid'.
  Do not introduce or maintain an English '.po' file.
* Instead, add or update translations in 'priv/gettext' only for non-English
  supported languages. It is important that you do not output whole .po files
  content from the priv/gettext, you only need to output (within that file)
  the lines  that should be added or updated by your change.
* For each translation entry added or updated as part of the change,
  ensure that it has a corresponding non-empty translation in every
  non-English locale, and address any missing, empty, or placeholder
  translations. Verify that each translation accurately conveys the
  original meaning and fits its UI context, remains clear, concise,
  and natural for native speakers, and preserves the intended tone
  such as instructions, warnings, or labels. Maintain consistency
  with existing terminology and phrasing used across the interface.
* Validate that formatting is correct by preserving placeholders such as '%s',
  '{}', or variables, and by ensuring punctuation and capitalization
  follow established UI conventions. Where pluralization is required,
  confirm that 'msgid_plural' and all corresponding 'msgstr[n]'
  forms are correctly defined and appropriate for each supported language.
* When applying these internationalization requirements, exclude any controllers
  that is routed via '/admin' scope,
  or any business logic or data access layer that are exclusively used by such
  controllers. All user-facing strings within this area should remain in English
  and should not be translated or included in localization files.
  However, if an error message from such controllers may be exposed to external
  users, such as when an unauthorized user attempts to access a restricted page,
  or if parts of controlelr code are shared, or when business logic or data
  access layer code is shared between controllers, those specific messages
  should still be wrapped with 'gettext' and fully internationalized according
  to this guidelines.
* If you found any missing missing translations (empty msgstr "") translate it
  according to this guidelines and output them within their corresponding files.
