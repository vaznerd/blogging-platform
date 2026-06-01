# Blog Platform Architecture

## Overview

A multi-user blogging platform where users write posts in Markdown and publish them under their own profile.

Example URLs:

```text
/u/zed
/u/zed/learning-go
/u/alice/linux-notes
```

---

# Core Principles

* Markdown is the source of truth.
* HTML is generated and stored when posts are created or updated.
* Public pages are accessible without authentication.
* API is versioned from day one.
* Usernames are unique.
* Slugs are unique per user.
* PostgreSQL is the primary database.

---

# URL Structure

## Public Routes

```text
/                          Homepage
/u/:username               User profile
/u/:username/:slug         Blog post
/tags/:tag                 Posts by tag
/search                    Search page
```

Examples:

```text
/
 /u/zed
 /u/alice
 /u/zed/learning-go
 /tags/golang
 /search?q=postgres
```

---

## Dashboard Routes

```text
/dashboard
/dashboard/posts
/dashboard/posts/new
/dashboard/posts/:id/edit
/dashboard/settings
```

---

## Admin Routes

```text
/admin
/admin/users
/admin/posts
/admin/comments
```

---

# API Structure

Base path:

```text
/api/v1
```

---

## Authentication

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh

GET  /api/v1/auth/me
```

---

## Users

```http
GET    /api/v1/users/:username
PATCH  /api/v1/users/me
DELETE /api/v1/users/me
```

---

## Posts

```http
GET    /api/v1/posts
POST   /api/v1/posts

GET    /api/v1/posts/:id
PATCH  /api/v1/posts/:id
DELETE /api/v1/posts/:id
```

---

## Comments

```http
GET    /api/v1/posts/:id/comments
POST   /api/v1/posts/:id/comments

PATCH  /api/v1/comments/:id
DELETE /api/v1/comments/:id
```

---

## Tags

```http
GET /api/v1/tags
GET /api/v1/tags/:tag
```

---

# User Roles

## Guest

Can:

* Read posts
* Read profiles
* Search posts

Cannot:

* Create posts
* Comment
* Edit content

---

## User

Can:

* Create posts
* Edit own posts
* Delete own posts
* Create comments
* Edit own comments
* Delete own comments

Cannot:

* Modify content owned by others

---

## Admin

Can:

* Delete any post
* Delete any comment
* Suspend users
* Access admin panel

---

# Database Schema

## users

```sql
id
username
email
password_hash

bio
avatar_url

role

created_at
updated_at
```

Constraints:

```sql
UNIQUE(username)
UNIQUE(email)
```

---

## posts

```sql
id
author_id

title
slug

markdown_content
html_content

status

created_at
updated_at
published_at
```

Status:

```text
draft
published
archived
```

Constraints:

```sql
UNIQUE(author_id, slug)
```

---

## comments

```sql
id

post_id
author_id

content

created_at
updated_at
```

---

## tags

```sql
id
name
```

Constraints:

```sql
UNIQUE(name)
```

---

## post_tags

```sql
post_id
tag_id
```

---

# Content Publishing Flow

## Create Post

User writes:

```markdown
# Learning Go

Hello world.
```

Flow:

```text
User submits post
        ↓
Generate slug
        ↓
Render Markdown → HTML
        ↓
Sanitize HTML
        ↓
Store Markdown
        ↓
Store HTML
        ↓
Publish
```

Result:

```text
/u/zed/learning-go
```

---

## Read Post

Flow:

```text
Request post
      ↓
Fetch post from database
      ↓
Verify status = published
      ↓
Return stored HTML
      ↓
Render in browser
```

---

# Ownership Rules

Posts:

```text
author_id == current_user_id
```

Required for:

* Edit post
* Delete post

Comments:

```text
author_id == current_user_id
```

Required for:

* Edit comment
* Delete comment

Admins bypass ownership checks.

---

# Search

Searchable fields:

* Post title
* Post content
* Tags

Example:

```text
/search?q=golang
```

Future:

* PostgreSQL Full Text Search

---

# Pagination

Supported on listing endpoints.

Examples:

```text
?page=1
?page=2
?limit=20
```

---

# Middleware

Global:

* Request logging
* Recovery middleware
* Authentication middleware
* Rate limiting
* Security headers
* CORS

---

# Project Structure

```text
cmd/
    api/

internal/
    auth/
    user/
    post/
    comment/
    tag/

    middleware/
    database/
    config/

migrations/

web/
    templates/
    static/

docs/
```

---

# Future Features

* Likes
* Bookmarks
* Follow users
* Notifications
* Reading history
* Scheduled publishing
* RSS feeds
* View counters
* Newsletter subscriptions
* User themes
* Dark mode

```
```
