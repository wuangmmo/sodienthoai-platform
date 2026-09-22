# Admin & Operations V7

V7 introduces an operations control plane separate from public phone lookup APIs.

Admin endpoints expose aggregate operational health and moderation queues. Administrative mutations remain protected by the admin bearer credential until the dedicated identity/RBAC layer is introduced.

Every future privileged action should write an immutable audit record containing actor, action, resource and metadata. Long-running maintenance work is represented as operational jobs so imports, reindexing and SEO maintenance can be observed independently from HTTP requests.
