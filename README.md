# ConnectVN — Modern Social Networking Platform

ConnectVN is a full-featured **Social Networking Platform** built as a graduation thesis project. It is designed around **Clean Architecture** principles and integrates modern technologies — **Golang, React, WebSocket, Redis, Kafka, and MinIO** — to deliver real-time interactions, high scalability, and production-grade performance.

---

## 1. Project Introduction

ConnectVN provides an interactive online environment where users can connect, share content, and communicate in real time. The system is divided into two main subsystems:

- **Client Interface** — A rich user-facing application for social interactions, content creation, and real-time messaging.
- **Admin Dashboard** — A powerful back-office panel for monitoring platform growth, moderating content, and managing user accounts.

The platform is containerized using **Docker Compose**, making it easy to spin up the entire stack — including PostgreSQL, Redis, Kafka, and MinIO — with a single command.

---

## 2. Key Features

### User System
- **Authentication & Security** — Register with email/password (hashed via Argon2id), JWT-based login, and OTP-based password reset via Gmail SMTP (OTP stored in Redis with a 5-minute TTL).
- **Profile Management** — View and update personal profiles including avatar, cover photo (uploaded to MinIO), bio, and social counters (followers/following/posts).

### Posts & Social Interactions
- **Post Creation** — Compose text posts with optional image attachments (stored on MinIO S3-compatible storage).
- **News Feed** — Paginated timeline displaying posts in reverse chronological order.
- **Like / Unlike** — Toggle likes; events are published to **Kafka** for asynchronous counter updates and notification generation.
- **Comments** — Add, edit, and delete comments. Post owners receive instant notifications on new comments.
- **Bookmarks** — Save and revisit posts from a personal reading list.
- **Follow / Unfollow** — Follow other users; follower/following counters update in real time.

### Real-Time Messaging
- **1-on-1 Chat** — Persistent WebSocket connections for instant message delivery. Messages are stored in PostgreSQL and broadcast across nodes via **Redis Pub/Sub**.
- **Online Status** — User presence is tracked in Redis upon WebSocket connection/disconnection.
- **Unread Indicators** — Messages are flagged `is_read = false` until the recipient opens the conversation.

### Notifications
- **Real-Time Push** — Instant in-app notifications delivered via WebSocket for likes, comments, and new followers.
- **Notification History** — Paginated list of past notifications with mark-as-read support (individual or bulk).

### Admin Dashboard
- **Platform Overview** — Aggregated stats: total users, total posts, and activity rates.
- **Growth Analytics** — Daily growth charts for new accounts and posts over the last 7 days.
- **User Management** — Paginated, searchable user list with Ban / Unban capability (enforced at JWT middleware level).
- **Content Moderation** — View all posts with like/comment counts; permanently delete violating content.

---

## 3. Tech Stack

### Backend & Framework
| Layer | Technology |
|---|---|
| Language | Go (Golang) 1.25.0 |
| Web Framework | Gin HTTP Framework v1.12.0 |
| Architecture | Clean Architecture |
| Database | PostgreSQL with GORM ORM |

### Real-Time & Asynchronous Processing
| Layer | Technology |
|---|---|
| WebSocket | Gorilla WebSocket v1.5.3 |
| In-Memory / Pub-Sub | Redis v9.18.0 — OTP storage, online presence, message routing |
| Message Queue | Apache Kafka — Async processing of Like/Comment events to reduce DB load |

### Storage & Security
| Layer | Technology |
|---|---|
| Object Storage | MinIO (S3-Compatible) — avatars, cover photos, post images |
| Password Hashing | Argon2id — industry-standard memory-hard hashing algorithm |
| Token Auth | JWT (JSON Web Tokens) — stateless authentication |
| Email | Gmail SMTP — OTP delivery for password reset |

### Infrastructure
| Layer | Technology |
|---|---|
| Containerization | Docker & Docker Compose |
| Database Migrations | Custom SQL migration scripts |
