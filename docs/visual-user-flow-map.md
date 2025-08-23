# 🗺️ **Visual User Flow Map: Secure Email System**

## 📊 **Complete System Flow Overview**

```mermaid
graph TB
    %% User Entry Points
    Start([User Enters System]) --> UserType{User Type?}
    
    %% User Type Decision
    UserType -->|Internal User| InternalFlow[Internal User Flow]
    UserType -->|External User| ExternalFlow[External User Flow]
    UserType -->|Administrator| AdminFlow[Administrator Flow]
    
    %% Internal User Flow
    InternalFlow --> Login[Login with Email + Password + TOTP]
    Login --> Dashboard[Dashboard Access]
    Dashboard --> Compose[Compose Email]
    Compose --> SecurityConfig[Configure Security Settings]
    SecurityConfig --> RecipientCheck{Recipient Type?}
    
    %% Recipient Decision
    RecipientCheck -->|Internal| InternalSend[Send to Internal User]
    RecipientCheck -->|External| ExternalSend[Create Secure Link]
    
    %% External User Flow
    ExternalFlow --> ReceiveEmail[Receive Email with Secure Link]
    ReceiveEmail --> ClickLink[Click Secure Link]
    ClickLink --> SecurityValidation[Security Validation Process]
    SecurityValidation --> ViewEmail[View Secure Email]
    ViewEmail --> ReplyOption{Want to Reply?}
    ReplyOption -->|Yes| SecureReply[Send Secure Reply]
    ReplyOption -->|No| EndExternal[End Session]
    
    %% Security Validation Subflow
    SecurityValidation --> PasswordCheck{Password Required?}
    PasswordCheck -->|Yes| EnterPassword[Enter Password]
    PasswordCheck -->|No| GeoCheck{Geolocation Restricted?}
    EnterPassword --> GeoCheck
    GeoCheck -->|Yes| LocationCheck[Validate Location]
    GeoCheck -->|No| TimeCheck{Time Lock Active?}
    LocationCheck --> TimeCheck
    TimeCheck -->|Yes| WaitTime[Wait for Unlock Time]
    TimeCheck -->|No| MFACheck{MFA Required?}
    WaitTime --> MFACheck
    MFACheck -->|Yes| EnterMFA[Enter MFA Code]
    MFACheck -->|No| AccessGranted[Access Granted]
    EnterMFA --> AccessGranted
    
    %% Admin Flow
    AdminFlow --> AdminLogin[Admin Login with Enhanced MFA]
    AdminLogin --> AdminDashboard[Admin Dashboard]
    AdminDashboard --> SystemMonitor[System Monitoring]
    SystemMonitor --> UserManagement[User Management]
    UserManagement --> AuditLogs[Audit Logs]
    
    %% Styling
    classDef userFlow fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef securityFlow fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef decisionFlow fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef adminFlow fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    
    class InternalFlow,ExternalFlow,Login,Dashboard,Compose,ReceiveEmail,ClickLink,ViewEmail,SecureReply userFlow
    class SecurityConfig,SecurityValidation,PasswordCheck,GeoCheck,TimeCheck,MFACheck,EnterPassword,LocationCheck,WaitTime,EnterMFA,AccessGranted securityFlow
    class UserType,RecipientCheck,ReplyOption,PasswordCheck,GeoCheck,TimeCheck,MFACheck decisionFlow
    class AdminFlow,AdminLogin,AdminDashboard,SystemMonitor,UserManagement,AuditLogs adminFlow
```

---

## 🔐 **Authentication Flow**

```mermaid
graph TD
    %% New User Registration
    NewUser[New User Registration] --> EmailInput[Enter Email Address]
    EmailInput --> PasswordInput[Create Password]
    PasswordInput --> TOTPSetup[Setup TOTP Authenticator]
    TOTPSetup --> EmailVerify[Email Verification]
    EmailVerify --> AccountActive[Account Activated]
    
    %% Existing User Login
    ExistingUser[Existing User Login] --> EmailLogin[Enter Email]
    EmailLogin --> PasswordLogin[Enter Password]
    PasswordLogin --> TOTPCode[Enter TOTP Code]
    TOTPCode --> LoginSuccess[Login Successful]
    LoginSuccess --> DashboardAccess[Dashboard Access]
    
    %% Security Checks
    PasswordLogin --> BruteForceCheck{Brute Force Check}
    BruteForceCheck -->|Too Many Attempts| AccountLocked[Account Locked]
    BruteForceCheck -->|Valid| TOTPCode
    
    TOTPCode --> IPCheck{IP Check}
    IPCheck -->|Suspicious| AdditionalVerification[Additional Verification]
    IPCheck -->|Valid| LoginSuccess
    
    %% Styling
    classDef registrationFlow fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef loginFlow fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    classDef securityCheck fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    
    class NewUser,EmailInput,PasswordInput,TOTPSetup,EmailVerify,AccountActive registrationFlow
    class ExistingUser,EmailLogin,PasswordLogin,TOTPCode,LoginSuccess,DashboardAccess loginFlow
    class BruteForceCheck,AccountLocked,IPCheck,AdditionalVerification securityCheck
```

---

## 📧 **Internal User Email Flow**

```mermaid
graph LR
    %% Email Composition
    ComposeEmail[Compose Email] --> RecipientInput[Enter Recipients]
    RecipientInput --> SubjectInput[Enter Subject]
    SubjectInput --> MessageInput[Write Message]
    MessageInput --> AttachmentUpload[Upload Attachments]
    AttachmentUpload --> SecuritySettings[Configure Security]
    
    %% Security Configuration
    SecuritySettings --> PasswordProtection{Password Protection?}
    PasswordProtection -->|Yes| SetPassword[Set Password]
    PasswordProtection -->|No| GeoRestriction{Geolocation Restriction?}
    SetPassword --> GeoRestriction
    
    GeoRestriction -->|Yes| SetGeoLocation[Set Allowed Locations]
    GeoRestriction -->|No| TimeLock{Time Lock?}
    SetGeoLocation --> TimeLock
    
    TimeLock -->|Yes| SetTimeLock[Set Unlock Time]
    TimeLock -->|No| MFARequired{MFA Required?}
    SetTimeLock --> MFARequired
    
    MFARequired -->|Yes| SetMFA[Configure MFA Type]
    MFARequired -->|No| AdditionalSecurity[Additional Security Settings]
    SetMFA --> AdditionalSecurity
    
    %% Send Process
    AdditionalSecurity --> SendEmail[Send Email]
    SendEmail --> RecipientCheck{Recipient Type?}
    RecipientCheck -->|Internal| DirectSend[Direct Send]
    RecipientCheck -->|External| CreateSecureLink[Create Secure Link]
    
    CreateSecureLink --> SendNotification[Send Notification Email]
    SendNotification --> ExternalRecipient[External Recipient Receives]
    
    %% Styling
    classDef compositionFlow fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef securityConfig fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef sendFlow fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    
    class ComposeEmail,RecipientInput,SubjectInput,MessageInput,AttachmentUpload compositionFlow
    class SecuritySettings,PasswordProtection,SetPassword,GeoRestriction,SetGeoLocation,TimeLock,SetTimeLock,MFARequired,SetMFA,AdditionalSecurity securityConfig
    class SendEmail,RecipientCheck,DirectSend,CreateSecureLink,SendNotification,ExternalRecipient sendFlow
```

---

## 🌐 **External User Secure Link Flow**

```mermaid
graph TD
    %% Receive and Access
    ReceiveLink[Receive Secure Link Email] --> ClickSecureLink[Click Secure Link]
    ClickSecureLink --> SecurityValidation[Security Validation Process]
    
    %% Security Validation Chain
    SecurityValidation --> CheckExpiration{Link Expired?}
    CheckExpiration -->|Yes| LinkExpired[Link Expired - Access Denied]
    CheckExpiration -->|No| CheckRevoked{Link Revoked?}
    
    CheckRevoked -->|Yes| LinkRevoked[Link Revoked - Access Denied]
    CheckRevoked -->|No| CheckPassword{Password Required?}
    
    CheckPassword -->|Yes| PasswordPrompt[Enter Password]
    PasswordPrompt --> ValidatePassword{Password Correct?}
    ValidatePassword -->|No| PasswordFailed[Password Failed]
    PasswordFailed --> AttemptLimit{Attempt Limit Reached?}
    AttemptLimit -->|Yes| LinkDestroyed[Link Destroyed]
    AttemptLimit -->|No| PasswordPrompt
    ValidatePassword -->|Yes| CheckGeoLocation{Geolocation Restricted?}
    
    CheckPassword -->|No| CheckGeoLocation
    
    CheckGeoLocation -->|Yes| ValidateLocation{Location Valid?}
    ValidateLocation -->|No| LocationBlocked[Location Blocked]
    ValidateLocation -->|Yes| CheckTimeLock{Time Lock Active?}
    
    CheckGeoLocation -->|No| CheckTimeLock
    
    CheckTimeLock -->|Yes| CheckTime{Current Time Valid?}
    CheckTime -->|No| TimeBlocked[Time Lock Active - Wait Required]
    CheckTime -->|Yes| CheckMFA{MFA Required?}
    
    CheckTimeLock -->|No| CheckMFA
    
    CheckMFA -->|Yes| MFAPrompt[Enter MFA Code]
    MFAPrompt --> ValidateMFA{MFA Valid?}
    ValidateMFA -->|No| MFAFailed[MFA Failed]
    ValidateMFA -->|Yes| AccessGranted[Access Granted]
    
    CheckMFA -->|No| AccessGranted
    
    %% Email Viewing
    AccessGranted --> CreateSession[Create Viewing Session]
    CreateSession --> SanitizeContent[Sanitize Email Content]
    SanitizeContent --> DisplayEmail[Display Secure Email]
    DisplayEmail --> AttachmentAccess{Attachments?}
    AttachmentAccess -->|Yes| DownloadAttachment[Download Attachments]
    AttachmentAccess -->|No| ReplyOption{Want to Reply?}
    DownloadAttachment --> ReplyOption
    
    %% Reply Process
    ReplyOption -->|Yes| ComposeReply[Compose Secure Reply]
    ReplyOption -->|No| EndSession[End Session]
    ComposeReply --> SendReply[Send Secure Reply]
    SendReply --> NewSecureLink[Create New Secure Link]
    NewSecureLink --> NotifySender[Notify Original Sender]
    
    %% Styling
    classDef accessFlow fill:#e1f5fe,stroke:#0277bd,stroke-width:2px
    classDef securityCheck fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef viewingFlow fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef replyFlow fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    
    class ReceiveLink,ClickSecureLink,CreateSession,SanitizeContent,DisplayEmail,EndSession accessFlow
    class SecurityValidation,CheckExpiration,CheckRevoked,CheckPassword,PasswordPrompt,ValidatePassword,CheckGeoLocation,ValidateLocation,CheckTimeLock,CheckTime,CheckMFA,MFAPrompt,ValidateMFA,AccessGranted securityCheck
    class AttachmentAccess,DownloadAttachment viewingFlow
    class ReplyOption,ComposeReply,SendReply,NewSecureLink,NotifySender replyFlow
```

---

## ⚙️ **Administrative Flow**

```mermaid
graph TB
    %% Admin Access
    AdminLogin[Admin Login] --> AdminDashboard[Admin Dashboard]
    
    %% System Management
    AdminDashboard --> SystemHealth[System Health Monitoring]
    SystemHealth --> PerformanceMetrics[Performance Metrics]
    PerformanceMetrics --> ErrorTracking[Error Tracking]
    ErrorTracking --> ResourceUsage[Resource Usage]
    
    %% User Management
    AdminDashboard --> UserManagement[User Management]
    UserManagement --> ViewUsers[View All Users]
    ViewUsers --> AccountStatus[Account Status Management]
    AccountStatus --> SecuritySettings[Security Settings Review]
    SecuritySettings --> AccessControl[Access Control]
    
    %% Security Administration
    AdminDashboard --> SecurityAdmin[Security Administration]
    SecurityAdmin --> AuditLogs[Audit Logs]
    AuditLogs --> SecurityIncidents[Security Incidents]
    SecurityIncidents --> ComplianceReports[Compliance Reports]
    ComplianceReports --> PolicyManagement[Policy Management]
    
    %% Enterprise Features
    AdminDashboard --> EnterpriseFeatures[Enterprise Features]
    EnterpriseFeatures --> QuotaManagement[Quota Management]
    QuotaManagement --> RetryLogic[Retry Logic]
    RetryLogic --> SystemResilience[System Resilience]
    
    %% Styling
    classDef adminFlow fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef systemFlow fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    classDef securityFlow fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef enterpriseFlow fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    
    class AdminLogin,AdminDashboard adminFlow
    class SystemHealth,PerformanceMetrics,ErrorTracking,ResourceUsage systemFlow
    class SecurityAdmin,AuditLogs,SecurityIncidents,ComplianceReports,PolicyManagement securityFlow
    class EnterpriseFeatures,QuotaManagement,RetryLogic,SystemResilience enterpriseFlow
```

---

## 🔒 **Security Features Flow Map**

```mermaid
graph LR
    %% Security Features Overview
    SecurityFeatures[Security Features] --> PasswordProtection[Password Protection]
    SecurityFeatures --> GeolocationVerification[Geolocation Verification]
    SecurityFeatures --> TimeBasedControls[Time-based Controls]
    SecurityFeatures --> MultiFactorAuth[Multi-Factor Authentication]
    SecurityFeatures --> ReadOnce[Read-once & Auto-destruct]
    SecurityFeatures --> RemoteRevocation[Remote Revocation]
    SecurityFeatures --> DecoyMessages[Decoy Messages]
    SecurityFeatures --> MetadataStripping[Metadata Stripping]
    SecurityFeatures --> TamperDetection[Tamper Detection]
    SecurityFeatures --> SelfDestruct[Self-destruct After Failed Attempts]
    SecurityFeatures --> EmailExpiration[Email Expiration]
    SecurityFeatures --> AuditLogging[Enhanced Audit Logging]
    
    %% Implementation Flow
    PasswordProtection --> Argon2Hashing[Argon2 Hashing]
    GeolocationVerification --> IPGeolocation[IP Geolocation]
    TimeBasedControls --> UnixTimestamp[Unix Timestamp Validation]
    MultiFactorAuth --> TOTPImplementation[TOTP Implementation]
    ReadOnce --> DatabaseDeletion[Database Deletion]
    RemoteRevocation --> LinkStatusManagement[Link Status Management]
    DecoyMessages --> ConditionalContent[Conditional Content Display]
    MetadataStripping --> ContentSanitization[Content Sanitization]
    TamperDetection --> ActivityMonitoring[Activity Monitoring]
    SelfDestruct --> AttemptTracking[Attempt Tracking]
    EmailExpiration --> TimeBasedDeletion[Time-based Deletion]
    AuditLogging --> ComprehensiveLogging[Comprehensive Logging]
    
    %% Styling
    classDef featureFlow fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef implementationFlow fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    
    class SecurityFeatures,PasswordProtection,GeolocationVerification,TimeBasedControls,MultiFactorAuth,ReadOnce,RemoteRevocation,DecoyMessages,MetadataStripping,TamperDetection,SelfDestruct,EmailExpiration,AuditLogging featureFlow
    class Argon2Hashing,IPGeolocation,UnixTimestamp,TOTPImplementation,DatabaseDeletion,LinkStatusManagement,ConditionalContent,ContentSanitization,ActivityMonitoring,AttemptTracking,TimeBasedDeletion,ComprehensiveLogging implementationFlow
```

---

## 📱 **User Journey Timeline**

```mermaid
gantt
    title Secure Email System User Journey Timeline
    dateFormat  YYYY-MM-DD
    section Internal User
    Registration           :done, reg, 2025-01-01, 1d
    First Login           :done, login, after reg, 1d
    Compose Email         :active, compose, after login, 2d
    Configure Security    :security, after compose, 1d
    Send Email           :send, after security, 1d
    
    section External User
    Receive Email        :receive, after send, 1d
    Click Secure Link    :click, after receive, 1d
    Security Validation  :validation, after click, 2d
    View Email          :view, after validation, 1d
    Reply (Optional)    :reply, after view, 1d
    
    section Administrator
    System Monitoring    :monitor, 2025-01-01, 7d
    User Management      :manage, 2025-01-02, 5d
    Security Review      :review, 2025-01-03, 3d
    Audit Logs          :audit, 2025-01-04, 4d
```

---

## 🎯 **Key Decision Points**

### **1. User Type Selection**
- **Internal User**: Full system access with all features
- **External User**: Secure link-based access only
- **Administrator**: System management and monitoring

### **2. Security Configuration (Optional)**
- **Password Protection**: Enable/disable with custom password
- **Geolocation Restrictions**: Country/city-level restrictions
- **Time-based Controls**: Expiration and time lock settings
- **Multi-Factor Authentication**: TOTP, SMS, or Email-based
- **Access Controls**: Read-once, auto-destruct, attempt limits

### **3. Recipient Type Detection**
- **Internal Recipients**: Direct email delivery
- **External Recipients**: Automatic secure link creation

### **4. Security Validation Chain**
- **Expiration Check**: Verify link hasn't expired
- **Revocation Check**: Verify link hasn't been revoked
- **Password Validation**: If password protection enabled
- **Geolocation Validation**: If location restrictions set
- **Time Lock Validation**: If time restrictions active
- **MFA Validation**: If multi-factor authentication required

### **5. Reply Decision**
- **Reply**: Create new secure link for ongoing conversation
- **No Reply**: End session and close secure link

---

## 📊 **Flow Statistics**

| Flow Type | Steps | Security Checks | User Types | Features |
|-----------|-------|-----------------|------------|----------|
| **Internal User** | 15+ | 3 | 1 | All 12 Security Features |
| **External User** | 20+ | 8 | 1 | Subject to Sender's Settings |
| **Administrator** | 12+ | 5 | 1 | System Management Tools |
| **Security Validation** | 8 | 8 | 2 | Real-time Enforcement |

---

## 🔄 **System Integration Points**

```mermaid
graph TB
    %% Core System Components
    Frontend[React Frontend] --> BackendAPI[Go Backend API]
    BackendAPI --> Database[(SQLite Database)]
    BackendAPI --> Storage[Cloudflare R2 Storage]
    
    %% Security Components
    BackendAPI --> Encryption[Post-Quantum Encryption]
    BackendAPI --> Authentication[JWT + TOTP Auth]
    BackendAPI --> Geolocation[IP Geolocation Service]
    
    %% External Services
    BackendAPI --> EmailService[Email Service]
    BackendAPI --> MFAService[MFA Service]
    BackendAPI --> AuditService[Audit Service]
    
    %% User Interfaces
    Frontend --> InternalUI[Internal User Interface]
    Frontend --> ExternalUI[External User Interface]
    Frontend --> AdminUI[Administrator Interface]
    
    %% Styling
    classDef coreSystem fill:#e3f2fd,stroke:#1565c0,stroke-width:2px
    classDef securitySystem fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef externalService fill:#e8f5e8,stroke:#2e7d32,stroke-width:2px
    classDef userInterface fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    
    class Frontend,BackendAPI,Database,Storage coreSystem
    class Encryption,Authentication,Geolocation securitySystem
    class EmailService,MFAService,AuditService externalService
    class InternalUI,ExternalUI,AdminUI userInterface
```

---

*This visual flow map provides a comprehensive overview of all user journeys through the Secure Email system, showing decision points, security validations, and system integrations.*
