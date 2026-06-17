# Please follow these guidelines when editing code

## Files

* Provide the complete, updated source code for every file you modify or add,
  including all relevant comments. Do not omit any part of these files.
* If your response includes files, format the output so that each file starts
  with the filename in **bold**, followed immediately by its content inside
  a fenced triple-backtick code block, without any extra text.
* If multiple files are provided, only include those that were actually changed.
  Do not include any files whose content remains unchanged.
* If a file must be removed from the project because it is no longer needed,
  output a separate line for each file to be deleted. Each line must begin
  with `DELETE`, followed by the filename in **bold**.
  For example: `DELETE **file-i-dont-need-anymore.txt**`

## Comments

* Avoid comments like "added this" or "changed that". These are only useful
  during a specific code change and provide little long-term value. This does
  not mean comments should be avoided. Instead, focus on writing comments that
  explain why something exists or how it works, in a way that remains useful
  over time.
* Keep existing comments unchanged unless you are modifying the code directly
  related to that comment.

* This instructions does not mean that you should output only the source code.
  You should still include your usual explanations of the changes or answers
  to my questions outside the code block. In some cases, you may answer
  a question by modifying the code itself, and that is also welcome.
  When appropriate, provide both the updated code and a brief explanation
  of what was changed and why.

## Output format

* Do not provide results in patch or diff format, even if such a format appears
  anywhere in the input.

## Approach to code style and refactoring

* When modifying code, strive to keep functions around 30-40 lines unless 
  complexity strictly requires more, extracting logic into sub-functions where
  beneficial. If code logic beign changed due to the current objective
  overlaps with existing logic elsewhere in the codebase, consider
  refactorig that shared logic into a common function; you are
  permitted to modify the unrelated sections of the codebase only as far as
  necessary to facilitate this extraction and ensure both areas use the new
  common function. Beyond this specific type of deduplication, do not
  perform any refactoring or changes to sections of the code not
  directly related to the current objective.
