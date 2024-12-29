# Concept sketch

## Goal

1. **Why do you want to develop your own app for newsletters?** 
   Newsletters are extremely confusing. 
   They are not prioritized according to my interests. 
   They are not grouped by topic. 
   They are difficult to filter. 
   It is difficult to keep track of which stories I have already read (since a
   newsletter often links several stories).

2. **Scope.** 
   The focus is clearly on newsletters -- both those that essentially present
   a story in themselves (e.g. a continuous text email), as well as those
   that contain several stories (e.g. an email with three separate stories), 
   and finally those that simply link several stories (e.g. with a teaser image,
   headline and teaser text).

3. **User group.** 
   The focus is initially on me as the user.

4. **Presentation goal.** 
   The user sees stories with a headline and short text, possibly also with an
   image. 
   These are clickable and lead to a website with the complete story. The 
   stories are prioritized according to the user's personal interests (e.g. 
   from historical usage behavior). The stories can be filtered by topic (e.g.
   displayed as tags).

5. **Data protection.** 
   The app processes personal data to a very limited extent (user credentials 
   in conjunction with individual user preferences). These must be protected 
   with the usual security measures (e.g. login with a strong password).

6. **End devices.** 
   Ideally, the app is implemented as a web app that can be used on any common
   platform (smartphone, tablet, desktop).

## System architecture

The app will consist of the following components.

* **Backend.**
  Lightweight.
  Will poll the email server in regular intervals to check for unknown 
  newsletters.
  Parses all incoming newsletters.
  Provides views for the UI.
  Provides a RESTful API for the UI.
  Integrates with one or multiple data storage solutions.

* **UI.**
  A web app that connects to the backend via RESTful API.

## Email integration

Considering three options:

1. The app pulls from the mailbox via POP3/IMAP credentials.
   Relatively simple to set up for the user and most flexible option.
   However, needs to deal with IMAP protocol specifics.
   Also needs to store the IMAP credentials.

2. The app offers an SMTP endpoint to receive forwarded emails.
   Complicated in development and operations.
   Simple to set up for the user, yet not very flexible.

3. The app leverages a 3rdparty service to receive emails.
   It accesses them through a RESTful API.
   Simple to set up for the user.
   Simple in development and operations, also because the 3rdparty service can
   cache the emails temporarily for multiple hours in case of operational
   issues.
   Mailgun has a free tier for up to 100 emails per day.

The 3rdparty service is tempting, but in the spirit of simple software with
little dependencies and orthogonal feature sets, the app will initially
integrate with an IMAP endpoint.

## Data storage

The app shall have no offline mode in the beginning. However, it needs its own
persistence to store the e-mails, store intermediate analysis results, and to 
learn from the users behavior.

It will not permanently store the entire raw email.
Instead, it will only keep what's required for later display by the user.

## Personalization

Groking through the flood of newsletters is no easy task these days.
Therefore, machine learning features for sorting and filtering the emails will 
be highly valuable to the user.
Hence the app will have such features in very early stages.

Yet, the top-most priority is properly parsing and displaying the stories from
the newsletters.

## User Experience

Minimal for starters, will refine later on.


