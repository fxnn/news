A program that reads downloaded e‑mail
data from a Maildir directory.

It processes all e-mails using a configured LLM tool,
and fixed prompts, in order to extract one or multiple
stories contained in the e-mail.

For every story, it extracts a headline, a short teaser,
and a URL which can be opened in the browser to read the
story.

It stores all found stories in a directory.

The program shall be a CLI application, written in Go.
As command line arguments, it will accept a mail directory
for the e-mails to consume, a story directory for the
stories to output, some kind of LLM configuration,
depending on the library used. More will be added as
needed.
