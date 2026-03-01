# URL Shortener Application Plan

This document outlines the idea and development plan for a URL shortener application.

## Initial idea

The initial idea for this project was to create a simple URL shortener application that registered users could use to shorten their long URLs. The application would allow users to log in with their Google, Facebook, or other third-party accounts (the exact list has not been finalized and is intended to cover as many users as possible) and manage their shortened URLs through a dashboard. The main requirements when choosing technologies were the following:

1. The initial version of the app should be hosted on a Linux server with 4GB of RAM on the cheapest DigitalOcean droplet.
2. The database should support SQL and be easy to scale in the future, including replication and sharding.
3. Both frontend and backend solutions should be lightning fast and easy to deploy on the same server. And neither of them should become a bottleneck for app performance, neither in the initial version nor in future versions as the app scales.
4. The app should support user registration and authentication through third-party providers (no passwords or sensitive information stored in the DB), the ability to add new third-party authentication credentials for the same user if necessary, without limiting the number of providers per user and accounts per provider, and the app should be able to handle the authentication flow for all supported providers without any issues. It should also implement CRUD for shortened URLs for authenticated users, with the option to allow users to set the shortcode themselves or have it generated automatically. It should also support an optional expiration date for the shortened URLs. The redirection should be lightning-fast, and the app should handle a large number of redirects without performance issues. Any performance bottlenecks should be easily identified and resolved by adding more server resources. But it is essential that the app uses resources as efficiently as possible. Simple statistics on redirects should be collected to set restrictions for free users in the future, if necessary, and to prevent DDoS attacks and other abuse of the service.
5. Code maintainability. The codebase should be clean, well-structured, strictly typed, and easy to understand. It is highly preferable to use an ORM to interact with the database and avoid writing raw SQL queries whenever possible. It is also desirable to use a code-first approach for database schema management and migrations. It is also desirable to have a way to automatically generate API calls and the data types necessary to make them on the frontend from the backend codebase, to avoid issues with data types and API endpoint mismatches. An OpenAPI specification for all API is a must-have and must be generated from the backend codebase. The codebase should also be well-documented and include comprehensive tests to ensure the app's reliability and maintainability.

# Development Plan

## 1. Backend (Go): API and Logic

- **Framework:** `Gin`
- **Database Interaction:** `pq` package
- **Authentication:** `go-oidc` or similar
- **URL Shortening Logic:** Base62 encoding, URL validation
- **API Endpoints:**
  - `/shorten`: Create a short URL record.
  - `/redirect/{shortcode}`: Redirect to the long URL.
  - `/user/urls`: List user's shortened URLs.
  - `/user/urls/{id}`: CRUD for specific user URLs.

## 2. Database (PostgreSQL): Schema

- **Users Table:**
  - `user_id` (SERIAL PRIMARY KEY)
  - `provider` (VARCHAR) - e.g., "google", "facebook", "github".
  - `provider_user_id` (VARCHAR UNIQUE)
- **ShortenedUrls Table:**
  - `url_id` (SERIAL PRIMARY KEY)
  - `user_id` (INTEGER REFERENCES Users(user_id))
  - `shortcode` (VARCHAR UNIQUE NOT NULL)
  - `longurl` (TEXT NOT NULL)
  - `created_at` (TIMESTAMP WITH TIME ZONE)

## 3. Frontend (Vue.js): User Interface

- **Framework:** Vue.js
- **UI Library:** Vue Material, Vuetify
- **Authentication:** Official SDKs, Vuex
- **Components:**
  - Login/Signup Page
  - URL Shortening Form
  - Dashboard

## Key Considerations & Next Steps:

- **Security:** Input validation, rate limiting.
- **Error Handling:** Comprehensive error handling.
- **Caching:** Redis for URL redirection performance.
- **Scalability Testing:** Load testing, performance optimization.
- **Database Indexing:** Indexes on `user_id` and `shortcode`.
