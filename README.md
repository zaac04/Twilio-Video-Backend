# 📡 Go Video Orchestration Service

## Overview

This is a scalable backend service written in **Go** that orchestrates **Twilio Video** for managing live video sessions. It allows multiple participants to generate secure tokens and join rooms, supports custom room configurations, and post-processes video recordings using **AWS Elemental MediaConvert**, with final assets stored in **Amazon S3**.

---

## ✨ Key Features

- 🔒 **Participant Token Generation**: Secure access tokens for each participant with identity and role.
- 🛠️ **Custom Room Parameters**: Set duration, video quality, region, and other room configurations.
- 🔄 **Twilio Webhook Listener**: React to room lifecycle and recording events.
- 🎞️ **Post-Processing via MediaConvert**: Convert raw video to formats like `.mp4`, `.hls`, etc.
- ☁️ **Cloud Storage**: Upload processed media to **AWS S3**.
- ⚡ Built using Go, Chi router, and AWS/Twilio SDKs.

---

## 🔧 Tech Stack

- **Language**: Go (v1.21+)
- **Framework**: [Chi Router](https://github.com/go-chi/chi)
- **Video**: [Twilio Video API](https://www.twilio.com/video)
- **Cloud**: AWS S3 + MediaConvert
- **Security**: Token-based identity with optional webhook signature verification

---

## 🚀 Getting Started

### 1. Clone the repo

```bash
git clone https://github.com/zaac04/Twilio-Video-Backend.git
cd go-video-service
go mod tidy


