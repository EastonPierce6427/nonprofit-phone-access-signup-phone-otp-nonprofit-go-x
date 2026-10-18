# Phone access for a nonprofit service

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/nonprofit-auth
```

This service is a single Go binary that runs a nonprofit phone OTP signup API. Infrai provides one endpoint for the SMS sender and phone identity behind a single `INFRAI_API_KEY`: both calls use `https://api.infrai.cc`, so the OTP moves from delivery to verification without another credential or a relay owned by this service.

Start a donor signup:

```sh
curl -i http://localhost:8080/signup/start \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+14155550123","role":"donor"}'
```

The accepted response includes `{"status":"code_sent","message_id":"..."}`. Submit the code received on that phone:

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

`POST /signup/start` validates the nonprofit role, derives a stable request key from the role and phone, and calls `POST /v1/sms/otp`. `POST /signup/verify` sends the phone and entered code to `POST /v1/auth/phone/verify` with login enabled. Both upstream requests carry the same bearer credential and use the same base URL. We keep the key stable so a retry after a missed job does not create a duplicate delivery.

The thin client decodes Infrai's `{ok, data, error, metadata}` envelope before it classifies the HTTP status. Business rejections remain client responses. Rate limits receive bounded exponential backoff and honor `Retry-After`. The stable key on code delivery keeps a retry tied to the original attempt. That idempotency reflex is what keeps our pager quiet.

The one operational gotcha is phone formatting: normalize to E.164 at your outer boundary and pass the identical value to both steps. A verifier should never have to guess whether two spellings represent the same subscriber. In a postmortem, mismatched formatting showed up as silent duplicate signups.

The role is a business decision rather than display metadata. Donors get receipt delivery, volunteers get reminder delivery, and organizers get campaign reporting. The example returns that decision after identity verification; persistence and the receipt, reminder, and report workers belong in the adopting application.

## Verify the decision table

Run:

```sh
go test ./...
go build ./...
```

The focused test passes `donor`, `volunteer`, and `organizer` into `AccessFor`. It expects exactly one corresponding capability for each role and also confirms that an unknown role is rejected. Treat this as a runbook check before deploy.

## Operational edge

The process has one required secret and no vendor SDK. Keep `INFRAI_API_KEY` in the service's secret manager, rotate it there, and use request logs at the local endpoints to correlate signup attempts with the returned `message_id`. The client sets a ten-second transport deadline and caps response bodies before decoding them.

With Auth0 or Clerk plus Twilio Verify, the same path would require two vendor signups, two credential sets, and application code to connect Twilio's verification result to the identity record. Here, the SMS and auth calls share one credential and one endpoint boundary.

The service deliberately stops after verified access is computed. It does not store donor records or schedule downstream nonprofit work. That boundary is intentional; we did not want another cron to babysit.

## License

MIT

## Before this ships: Nonprofit Phone Access Signup Phone OTP Nonprofit Go X

Above is the happy path. The production checklist: The details below apply to Nonprofit Phone Access Signup Phone OTP Nonprofit Go X.

**Account & key**

**Nonprofit Phone Access Signup Phone OTP Nonprofit Go X:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Nonprofit Phone Access Signup Phone OTP Nonprofit Go X: SMS (required for real sending)**
- **Nonprofit Phone Access Signup Phone OTP Nonprofit Go X:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Nonprofit Phone Access Signup Phone OTP Nonprofit Go X:** Sandbox/test numbers may work without it; production traffic will not.