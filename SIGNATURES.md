# E2E Package Method Signatures - Canonical Definitions

## Function: DecryptMessage

### Canonical Signature (Client)
```go
func (c *Client) DecryptMessage(message *Message, senderPublicKey []byte) ([]byte, error)
```

### Canonical Signature (CryptoProvider)
```go
func (cp *CryptoProvider) DecryptMessage(envelope *Envelope, recipientPrivateKey []byte, senderPublicKey []byte) ([]byte, error)
```

### Call Sites with Issues
- `loadtest.go:552`: `us.client.DecryptMessage(message)` - Missing `senderPublicKey` parameter
- `security_test_suite.go:733`: `sts.ProtocolAnalyzer.client.DecryptMessage(message)` - Missing `senderPublicKey` parameter

### Fix Needed
Update call sites to include the required `senderPublicKey []byte` parameter.

---

## Function: EncryptThreadMessage

### Canonical Signature
```go
func (c *Client) EncryptThreadMessage(plaintext []byte, thread *Thread) (*Message, error)
```

### Call Sites with Issues
- `security_test_suite.go:448`: `sts.ProtocolAnalyzer.client.EncryptThreadMessage(thread.ID, msg)` - Wrong parameter order and types

### Fix Needed
Update call site to use `(plaintext []byte, thread *Thread)` instead of `(thread.ID string, msg []byte)`.

---

## Function: DecryptThreadMessage

### Canonical Signature
```go
func (c *Client) DecryptThreadMessage(message *Message, thread *Thread) ([]byte, error)
```

### Call Sites with Issues
- `benchmark.go`: Various calls with incorrect signatures

### Fix Needed
Update call sites to use `(message *Message, thread *Thread)` signature.

---

## Function: AuditLog

### Canonical Signature
```go
func (kt *KeyTransparency) AuditLog(fromIndex, toIndex int) ([]*AuditResult, error)
```

### Call Sites with Issues
- `benchmark.go:461`: Error message suggests wrong return type handling

### Fix Needed
Ensure call sites properly handle the `([]*AuditResult, error)` return type.

---

## Function: SignMessage (Missing)

### Issue
- `security_test_suite.go:624`: `sts.CryptoValidator.cryptoProvider.SignMessage` undefined

### Fix Needed
Either implement SignMessage on CryptoProvider or remove the call.

---

## Function: VerifySignature (Missing)

### Issue
- `security_test_suite.go:629`: `sts.CryptoValidator.cryptoProvider.VerifySignature` undefined

### Fix Needed
Either implement VerifySignature on CryptoProvider or remove the call.

---

## Function: GetKeyInfo

### Issue
- `loadtest.go:564`: `keyInfo.Algorithm` undefined (type map[string]interface{} has no field or method Algorithm)

### Fix Needed
Update to use map access: `keyInfo["Algorithm"]` instead of `keyInfo.Algorithm`.

---

## Function: SequenceNumber (Missing Field)

### Issue
- `security_test_suite.go:764`: `message.Envelope.SequenceNumber` undefined

### Fix Needed
Either add SequenceNumber field to Envelope struct or remove the access.

---

## Summary of Required Fixes

1. **DecryptMessage calls**: Add missing `senderPublicKey []byte` parameter
2. **EncryptThreadMessage calls**: Fix parameter order and types
3. **DecryptThreadMessage calls**: Update to use correct signature
4. **AuditLog calls**: Ensure proper return type handling
5. **Missing methods**: Implement or remove SignMessage/VerifySignature calls
6. **Map access**: Fix keyInfo.Algorithm to keyInfo["Algorithm"]
7. **Missing fields**: Add SequenceNumber to Envelope or remove access
8. **Unused variables**: Remove or use declared variables
