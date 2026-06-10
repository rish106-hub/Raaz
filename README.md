# Raaz

**Structured conversations for meaningful human connection**

Raaz is an anonymous, prompt-driven conversation platform built for Indian Gen Z. Unlike dating apps, random chat, or AI companions, Raaz helps strangers have emotionally honest conversations through shared prompts and structured interaction.

No profiles. No photos. No follower counts. Just real people, real conversations.

---

## Why Raaz?

**The Problem:** Young Indians are more connected digitally than ever—yet 70% of Gen Z report feeling isolated. 43% of urban Indians experience loneliness. Existing platforms fail because dating apps demand romantic intent, random chat apps lack trust, AI companions aren't human, and social media is performance-driven.

**The Solution:** Start conversations from shared emotional context, not introductions. When two people begin with the same vulnerable prompt, connection happens faster and deeper.

---

## How It Works

### Daily Flow
1. **Receive a Prompt** — One daily emotional prompt at 8 AM IST
   - Categories: Identity, Ambition, Relationships, Society, Regret, Gratitude
2. **Select Echo** — Tap to enter the matching pool
3. **Get Matched** — Paired anonymously with another user in your city/age bracket
4. **Conversation** — 20-minute chat window with extension option
5. **Disappear** — Content deleted after 48 hours

### Key Product Principles
- **Anonymous** — No names, photos, profiles visible during conversation
- **Human-to-human** — Real people only, no bots or AI
- **Time-bound** — 20 minutes + optional 10-minute extension
- **Emotionally safe** — Moderation, crisis support, reportable users
- **Ephemeral** — Conversations vanish in 48 hours (optional save to Vault for Deep members)

---

## Product Features

### Core
- **Daily Prompt System** — Rotating emotional prompts across six categories
- **Smart Matching** — City + age bracket matching with national fallback after 15 minutes
- **Minimal Chat Interface** — Text only (no GIFs, stickers, reactions)
- **Contact Exchange** — Mutual consent required to share Raaz handles
- **Moderation & Safety** — Real-time abuse detection, crisis support resources, strike system

### Freemium Tiers
| Feature | Free | Plus (₹99/mo) | Deep (₹199/mo) |
|---------|------|---------------|----------------|
| Conversations/day | 2 | Unlimited | Unlimited |
| Prompt Archive | ✗ | ✓ | ✓ |
| Category Selection | ✗ | ✓ | ✓ |
| Voice Mode | ✗ | ✗ | ✓ |
| Vault (Save Messages) | ✗ | ✗ | ✓ |

---

## Target Users

### Aakash (21, Bengaluru)
Engineering student in hostel with friends but lacking emotional depth. Needs judgment-free space to be heard.

### Shreya (23, Mumbai)
Early-career professional, newly relocated. Seeking low-pressure human connection without romantic expectation.

### Rohan (22, Delhi NCR)
Freelancer with large online presence and social anxiety. Wants depth, not endless shallow interactions.

---

## Technical Architecture

### Frontend
- **Android** (primary)
- **iOS**

### Backend
- **Go microservices** (core API, matching engine, moderation)
- **Node.js** (support services)

### Infrastructure
| Layer | Technology |
|-------|-----------|
| **Database** | PostgreSQL |
| **Session State** | Redis |
| **Real-time Messaging** | WebSockets |
| **Notifications** | Firebase Cloud Messaging + APNs |
| **Content Delivery** | Cloudflare CDN |

### Security & Privacy
- **Encryption:** TLS 1.3 in transit, encrypted storage at rest
- **Key Rotation:** Daily
- **No Ad Tracking:** Zero external tracking SDKs
- **Data Sales:** None
- **Compliance:** DPDP Act 2023, explicit consent flows, data deletion support

---

## Success Metrics

### North Star
**Weekly Active Conversationalists (WAC)** — Users completing ≥1 full conversation per week

### Supporting Metrics
- Conversation Extension Rate: >35%
- Contact Exchange Rate: >12%
- D7 Retention: >40%
- D30 Retention: >22%
- Prompt → Match Conversion: >55%
- Early Abandonment: <20%
- Abuse Reports: <2 per 1,000 sessions

---

## Launch Timeline

### Phase 1: Growth (Months 0–9)
- Free tier only
- Launch cities: Bengaluru, Mumbai, Delhi NCR
- Goal: 100,000 MAU

### Phase 2: Freemium (Months 9–18)
- Plus & Deep tier introduction
- Goal: ₹60–80 lakh ARR, 30,000 paying subscribers

### Phase 3: B2B (Month 18+)
- College & corporate partnerships
- Wellness platform integrations
- Annual pricing: ₹50k–₹2.00L per institution

---

## Development Setup

Coming soon.

---

## Security & Moderation

### Abuse Detection
Real-time detection for:
- Harassment
- Explicit content
- Self-harm indicators

### Enforcement
| Strike | Action |
|--------|--------|
| 1 | Warning |
| 2 | 7-day ban |
| 3 | Permanent ban |

### Crisis Support
Self-harm indicators trigger:
- Conversation pause
- Emergency helpline resources
- Reporting to moderation team

---

## Privacy

- Conversations deleted after 48 hours (no permanent history)
- Optional Vault for Deep members (user-controlled saves only)
- No data sold to third parties
- DPDP Act 2023 compliant
- Export/delete your data anytime

---

## Contributing

We're hiring. [Jobs Page] (coming soon)

---

## Contact

- **Email:** hello@raaz.app (coming soon)
- **Twitter:** [@raaz_app](https://twitter.com/raaz_app) (coming soon)

---

## License

Proprietary — All rights reserved.

---

**Tagline:** *Some conversations are too real for Instagram.*
