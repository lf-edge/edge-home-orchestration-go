# Security policy

### Contents
1. [Introduction](#1-introduction)
2. [Vulnerabilities](#2-vulnerabilities)  
    2.1 [Supported Versions](#21-supported-versions)  
    2.2 [Vulnerability Report](#22-vulnerability-report)  
    2.3 [Security Disclosure](#23-security-disclosure)  
3. [Security requrements](#3-security_requirements)
4. [Security Software life cycle processes](#3-security-software-life-cycle-processes)

---

## 1. Introduction

This document describes the sequence of actions when the vulnerability is founded, the product version that fixes them, security requirements, as well as the required process of developing a secure code.
 
 > To view the Items 3 & 4, you need to install a `plantuml` extension for your browser.

---

## 2. Vulnerabilities

### 2.1 Supported Versions

We are releasing patches to eliminate vulnerabilities, you can see below:

| Version     | Supported by | Edge-Orchestration | 3-rd party component                           |
| ----------- | ------------ | ------------------ | ---------------------------------------------- |
| 1.0.0       | N/A          |                    |                                                |
| 1.1.0       | Fixed        |                    | CVE-2020-15257, CVE-2021-32760, CVE-2021-41103 |
| 1.1.1       | Fixed        |                    | CVE-2021-41190                                 |
| 1.1.4       | Fixed        | CVE-2006-4624      |                                                |
| 1.1.6       | Fixed        | CVE-2006-4624      |                                                |
| 1.1.8       | Fixed        |                    | CWE-843                                        |
| 1.1.9       | Fixed        |                    | CVE-2022-23648                                 |
| 1.1.11      | Fixed        |                    | CVE-2021-3121                                  |

### 2.2 Vulnerability Report

The Edge-Orchestration team assigns the highest priority to all security bugs in Edge-Orchestration project. We appreciate your efforts and
responsible disclosure of information to eliminate vulnerabilities.

Please report security bugs by emailing the Security Issue Review (SIR) team at: homeedge-security-issues@lists.lfedge.org marked "SECURITY".
Our team will confirm your request and within 1 week will try to prepare recommendations for elimination. Our team will keep you updated on the progress towards the fix until the full announcement of the patch release. During this process, the edge-orchestration team may request additional information or guidance.


### 2.3 Security Disclosure

When a security group receives a security error report as previously mentioned, it is assigned the highest priority and the person in charge. This person will coordinate the patch and release process,
including the following steps:

  * Confirm the problem and identify the affected versions.
  * Check the code to find any similar problems.
  * Prepare fixes for all releases still in maintenance. These fixes will
    released as quickly as possible.

We suggest the following format when disclosing vulnerabilities:

  * Your name and email.
  * Include scope of vulnerability. Let us know who could use this exploit.
  * Document steps to identify the vulnerability. It is important that we can reproduce your findings. 
  * How to exploit vulnerability, give us an attack scenario.

---

## 3. Security requrements

```plantuml
@startuml

left to right direction
usecase "Security requirements"     #palegreen;line:black
usecase Confidentiality     as Co   #lightblue;line:black
usecase Integrity           as In   #lightblue;line:black
usecase Availability        as Av   #lightblue;line:black
usecase "Access control"    as Ac   #lightblue;line:black
usecase Identification              #lightblue;line:black
usecase Authentication              #lightblue;line:black
usecase Authorization               #lightblue;line:black
usecase Non                         #lightblue;line:black as "Non-public data 
    is kept confidential"
usecase "User privacy maintaned"    #lightblue;line:black
usecase "All data is confidential"  #lightblue;line:black
usecase "HTTPS: data in motion"     #lightblue;line:black
usecase "Authorization via GITHUB"  #lightblue;line:black
usecase Dtm                         #lightblue;line:black as "Data modification
    requires authorization"
usecase "Multiple backups"          #lightblue;line:black
usecase "Rerstore after DDoS"       #lightblue;line:black


(Security requirements) <-- (Co)    #line:black;line.bold
(Security requirements) <-- (In)    #line:black;line.bold
(Security requirements) <-- (Av)    #line:black;line.bold
(Security requirements) <-- (Ac)    #line:black;line.bold

(Ac) <-- (Identification)           #line:black
(Ac) <-- (Authentication)           #line:black
(Ac) <-- (Authorization)            #line:black
(Co) <-- (User privacy maintaned)   #line:black
(Co) <-- (Non)                      #line:black
(Co) <-- (All data is confidential) #line:black
(Co) <-- (HTTPS: data in motion)    #line:black
(In) <-- (HTTPS: data in motion)    #line:black
(In) <-- (Authorization via GITHUB) #line:black
(In) <-- (Dtm)                      #line:black
(Av) <-- (Multiple backups)         #line:black
(Av) <-- (Rerstore after DDoS)      #line:black

@enduml
```
---

## 4. Security Software life cycle processes
```plantuml
@startuml

left to right direction
usecase SSLCP           #palegreen;line:black   as  "Security Software
    life cycle processes"
usecase "Certification & Controls"      as CC       #lightblue;line:black
usecase CBPB            #lightblue;line:black   as  "CII Best 
    Practices badge"
usecase "OpenSSF Score Card"            as OSSFSC   #lightblue;line:black
usecase "Security in maintenance"       as SM       #lightblue;line:black
usecase ADPV            #lightblue;line:black   as  "Auto-detect publicy
    vulnerabilities"
usecase "Rapid update"                  as RU       #lightblue;line:black
usecase KDKDSS          #lightblue;line:black   as  "Key developers know how to
    develop secure software"
usecase "Infrastructure management"     as IM       #lightblue;line:black
usecase DTEPA           #lightblue;line:black   as  "Development & test
    environments protected
    from attack"
usecase CIATEP          #lightblue;line:black   as  "CI automated test
    environment does not have
    protected data"
usecase SIV             #lightblue;line:black   as  "Security in integration
    & verification"
usecase "Style checking tools"          as SCT      #lightblue;line:black
usecase SCWA            #lightblue;line:black   as  "Source code
    weakness analyzer"
usecase FLOSS           #lightblue;line:black
usecase "Negative Testing"              as NT       #lightblue;line:black
usecase UTC             #lightblue;line:black   as  "Unit Test
    coverage >75%"
usecase "Security in design"            as SD       #lightblue;line:black
usecase "Simple design"                 as SID      #lightblue;line:black
usecase "Memory-safe languages"         as MSL      #lightblue;line:black
usecase SDISS           #lightblue;line:black   as  "Secure disign
    includes S&S"


(SSLCP) <-- (CC)                    #line:black;line.bold
(SSLCP) <-- (SM)                    #line:black;line.bold
(SSLCP) <-- (KDKDSS)                #line:black;line.bold
(SSLCP) <-- (SIV)                   #line:black;line.bold
(SSLCP) <-- (IM)                    #line:black;line.bold
(SSLCP) <-- (SD)                    #line:black;line.bold

(CC)    <-- (CBPB)                  #line:black
(CC)    <-- (OSSFSC)                #line:black
(SM)    <-- (ADPV)                  #line:black
(SM)    <-- (RU)                    #line:black
(IM)    <-- (DTEPA)                 #line:black
(IM)    <-- (CIATEP)                #line:black
(SIV)   <-- (SCT)                   #line:black
(SIV)   <-- (SCWA)                  #line:black
(SIV)   <-- (FLOSS)                 #line:black
(SIV)   <-- (NT)                    #line:black
(SIV)   <-- (UTC)                   #line:black
(SD)    <-- (SID)                   #line:black
(SD)    <-- (MSL)                   #line:black
(SD)    <-- (SDISS)                 #line:black

@enduml
```
---
