# 🛠️ Employee PR Tracker

This project is designed to help track and manage GitHub Pull Requests (PRs) submitted by employees for performance review and accountability.

## 📌 Project Overview

The system allows employees to submit their GitHub PR links through a frontend. The backend then:

1. **Stores PR details** such as title, description, status, lines changed, and file changes in a database.
2. **Fetches PR data** using the GitHub REST API when a "Generate" button is clicked.
3. **Sends a summary mail** to a manager when the "Send Mail" button is clicked from the frontend.

## 👨‍💻 Key Features

- Store unique employee information with status
- Store and update pull request details
- Automatically fetch:
  - PR title and description
  - Status (open, closed, merged, draft)
  - Files changed
  - Lines added/removed
  - Commit count
- GitHub integration using Personal Access Token
- Clean and normalized database design
- Mail sending feature with formatted PR summary

## 🗃️ Tech Stack

- Golang (Go)
- GORM for ORM
- PostgreSQL (or any SQL DB)
- GitHub REST API
- SMTP / Email service (for mail feature)
- JSON API for frontend communication

