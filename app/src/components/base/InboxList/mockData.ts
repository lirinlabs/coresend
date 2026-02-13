import type { Email } from "./types"

export const mockEmails: Email[] = [
  {
    id: "1",
    from: "noreply@service.io",
    subject: "Verification Code: 847291",
    body: `Your verification code is: 847291

This code expires in 10 minutes.

If you did not request this code, please ignore this email.

—
Automated message from Service.io`,
    timestamp: new Date(Date.now() - 1000 * 60 * 5),
    ttl: "23h 55m",
  },
  {
    id: "2",
    from: "security@platform.dev",
    subject: "New login detected from unknown device",
    body: `A new login was detected on your account.

Device: Unknown
Location: [REDACTED]
Time: ${new Date().toISOString()}

If this was not you, please secure your account immediately.

—
Security Team`,
    timestamp: new Date(Date.now() - 1000 * 60 * 30),
    ttl: "23h 30m",
  },
  {
    id: "3",
    from: "newsletter@crypto.news",
    subject: "[WEEKLY] Market Update - Week 48",
    body: `WEEKLY MARKET DIGEST

— BTC: $43,291 (+2.4%)
— ETH: $2,847 (+1.8%)
— SOL: $98.42 (+5.2%)

Top Stories:
1. New regulatory framework proposed
2. DeFi TVL reaches new highs
3. Layer 2 adoption accelerates

Read more at crypto.news/weekly`,
    timestamp: new Date(Date.now() - 1000 * 60 * 120),
    ttl: "21h 00m",
  },
]
