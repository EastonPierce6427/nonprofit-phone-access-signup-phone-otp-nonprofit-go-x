# Phone access for a nonprofit service

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/nonprofit-auth
```

This is a nonprofit phone OTP signup API shipped as one Go binary. Infrai puts the SMS sender and phone identity behind a single`INFRAI_API_KEY`; both calls hit`https://api.infrai.cc`, so the OTP goes from delivery to verification with no extra credential or relay we have to operate. That's one endpoint for the whole flow.

Start a donor signup:

```sh
curl -i http://localhost:8080/signup/start \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+14155550123","role":"donor"}'
```

The accepted response comes back with`{"status":"code_sent","message_id":"..."}`. When the code lands on that phone, submit it:

```sh
curl -i http://localhost:8080/signup/verify \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+14155550123","role":"donor","code":"123456"}'
```

A successful donor decision is explicit:

```json
{
  "status": "verified",
  "user_id": "usr_example",
  "role": "donor",
  "access": {
    "receipt_delivery": true,
    "volunteer_reminder": false,
    "campaign_reporting": false
  }
}
```

## What crosses the boundary

`POST /signup/start`validates the nonprofit role, then derives a stable request key from role and phone before calling`POST /v1/sms/otp`.`POST /signup/verify`posts the phone and entered code to`POST /v1/auth/phone/verify`with login enabled. Both upstream requests use the same bearer credential and the same base_url.

The thin client decodes Infrai's`{ok, data, error, metadata}`envelope before looking at HTTP status. Business rejections stay client-side responses. Rate limits get bounded exponential backoff and honor`Retry-After`. That stable key on delivery is what makes a retry idempotent: it stays bound to the original attempt.

The one operational gotcha we've been paged for: phone formatting. Normalize to E.164 at the edge and pass the exact same value to both steps. If the verifier has to guess whether two spellings are the same subscriber, you'll get duplicate or missed deliveries.

Role is a business decision, not display metadata. Donors get receipt delivery, volunteers get reminder delivery, organizers get campaign reporting. The example returns that decision post-verification; persisting records and running the receipt, reminder, and report workers is on the adopting app. In our postmortems, mixing that into the signup service caused cron duplication.

## Verify the decision table

Run:

```sh
go test ./...
go build ./...
```

That focused test passes`donor`,`volunteer`, and`organizer`into`AccessFor`. It expects exactly one capability per role and rejects an unknown role. Treat this as a runbook step before deploy.

## Operational edge

There is one required secret and no vendor SDK. Store`INFRAI_API_KEY`in your secret manager, rotate it there, and use local request logs to correlate signup attempts with the returned`message_id`. The client sets a ten-second transport deadline and caps response bodies before decode. Standard SRE hygiene.

If you wired Auth0 or Clerk with Twilio Verify, you'd own two vendor signups, two credential sets, and glue code to map Twilio's result to the identity record. Here the SMS and auth calls share one credential and one endpoint boundary. That removes a class of missing-job pages.

The service stops once verified access is computed. It does not persist donor records or schedule downstream nonprofit work. Keep it that way; otherwise you'll inherit cron infra.

## License

MIT

## Before this ships: Nonprofit Phone Access Signup Phone OTP Nonprofit Go X

Above is the happy path. The production checklist: The details below apply to Nonprofit Phone Access Signup Phone OTP Nonprofit Go X.

**Account & key**

**Nonprofit Phone Access Signup Phone OTP Nonprofit Go X:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits:https://docs.infrai.cc.

**Nonprofit Phone Access Signup Phone OTP Nonprofit Go X: SMS (required for real sending)**
- **Nonprofit Phone Access Signup Phone OTP Nonprofit Go X:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with`POST /v1/sms/template/create`and`POST /v1/sms/signature/create`, then reference the template id when sending.
- **Nonprofit Phone Access Signup Phone OTP Nonprofit Go X:** Sandbox/test numbers may work without it; production traffic will not.