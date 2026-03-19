---
agent: Ask
model: Claude Sonnet 4.6 (copilot)
name: security-audit
description: Use this prompt to find security vulnerabilities in the codebase.
---

Perform a security audit of the codebase to detect any potential security vulnerabilities in this project. Focus on common security issues such as SQL injection, cross-site scripting (XSS), insecure authentication, and any other vulnerabilities that could be exploited by attackers. Consider OWASP Top Ten as a reference for common web application security risks. Review the code for best practices in secure coding and provide recommendations for any identified issues.

Output your findings in as a markdown formatted table with the following columns: "ID", "Severity", "Issue", "File Path", "Line Number(s)" and "Recommendation". The "ID" should start at 1 and increment for each issue found. The "File Path" should be relative to the repository root and it should be an actual link to the file on disk.

Limit your output to the top 10 most critical issues found, sorted by severity (Critical, High, Medium, Low). If more than 10 issues are found, provide a summary of the additional issues at the end of the table. If no issues are found, output "No security vulnerabilities detected."
