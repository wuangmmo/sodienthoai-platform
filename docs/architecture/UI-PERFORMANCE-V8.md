# UI/UX & Performance V8

The public experience is intentionally lookup-first: one primary action, clear verification language and no visual treatment that implies an unverified number is trustworthy.

The interface is responsive without external font or UI dependencies. System fonts avoid render-blocking font downloads. The home route contains a small amount of static content and the phone result route renders server-side metadata and structured data.

Delivery enables compression, HSTS, frame protection, MIME sniff protection and a restrictive permissions policy. Error and not-found states are explicit so transient API failures are not confused with missing phone records.

Before production, validate Core Web Vitals against the deployed infrastructure because database, Redis, OpenSearch, CDN and geographic latency cannot be accurately benchmarked from CI alone.
