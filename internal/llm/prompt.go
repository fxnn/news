package llm

import "fmt"

// extractionPromptTemplate is the prompt sent to the LLM to extract stories
// from a newsletter email. It contains two format verbs: subject and body.
const extractionPromptTemplate = `Your task is to extract news stories from an email. But first, decide whether this email is a CONTENT NEWSLETTER or not.

A content newsletter curates links to external articles, blog posts, podcasts, videos, repos, etc. — it points readers to content hosted elsewhere.

These are NOT content newsletters — return {"stories": []} for them:
- Marketing emails from brands (e.g. adidas, Amazon, Nike) promoting their own products or services
- Transactional emails (order confirmations, shipping updates, account notifications)
- Promotional emails with shopping links, discount codes, or product showcases
- Emails where all links point back to the sender's own website/shop/app

Subject: %s

Body:
%s

Return a JSON object with this exact structure:
{
  "stories": [
    {
      "headline": "Story headline",
      "teaser": "Content type prefix + reused summary if present; otherwise 2-4 sentences",
      "url": "https://example.com/article"
    }
  ]
}

CRITICAL INSTRUCTIONS:
- First decide: is this a content newsletter? If not, return {"stories": []}
- If it IS a content newsletter, extract ALL stories — not just the ones mentioned in the subject line
- Read through the ENTIRE email systematically from top to bottom
- Each story with a unique URL should be included
- Do NOT limit yourself to only a few stories - extract as many as exist

FORMATTING RULES:
- Write the headline and teaser in the same language as the original email
- Keep headlines SHORT: maximum 5-8 words
- Always start the teaser with a short content type label (1-2 words) followed by a period, e.g. "Article.", "Blog post.", "Podcast.", "Video.", "LinkedIn Post.", "GitHub Repo.", "Research Paper.", "News.", "Tutorial.", "Talk.", "Tool."
- If the newsletter already contains a summary paragraph describing the linked content, reuse that summary word-for-word after the content type prefix, regardless of length
- Otherwise, write teasers of 2-4 sentences. Prefer longer, more informative summaries over short ones.
- Each story MUST have a unique URL link to the actual article
- If there is only one URL in the email, create only one story
- Separate stories should have separate URLs - do not create multiple stories for a single URL

WHAT TO EXTRACT:
- Each story should be a MAIN article/post/resource being featured in the newsletter
- Extract the primary link for each distinct story/article
- Stories are typically presented as separate entries with their own headline and description

EXCLUSION RULES (apply these BEFORE adding any story):
- NEVER extract newsletter boilerplate. If a story is about the newsletter itself rather than external content, EXCLUDE it. Examples: "Werbung abbestellen", "Datenschutzinformationen", "Datenschutz", "Impressum", "Abmelden", "Unsubscribe", "Manage preferences", "Terms of Service", "Privacy Policy", "Cookie Policy". This applies in ALL languages.
- Exclude order links, shopping links, or any paid content
- Exclude sponsored content, advertisements, and promotions. A story is sponsored when it is LABELED as such by the newsletter — look for markers like "(Sponsor)", "Sponsored", "Ad", "Partner Post", "Promoted", "Brought to you by", "In partnership with" used as labels near the headline or as section headers. Do NOT exclude articles that merely discuss topics like advertising, partnerships, or affiliate programs as editorial content.
- Exclude giveaways, sweepstakes, contests, and raffles (Gewinnspiel, Verlosung, etc.) — these are promotions, not news
- Exclude social media links (follow us, share, tweet)
- Exclude footnote links, reference links, and citation links within story text
- Exclude "read more", "learn more", or supplementary links that are part of an existing story
- Only include actual news stories or articles with readable content
- If there are no valid stories with URLs, return {"stories": []}
`

func buildPrompt(subject, body string) string {
	return fmt.Sprintf(extractionPromptTemplate, subject, body)
}
