# Problem-Domain Relationships

**Generated:** ${new Date().toISOString()}
**Relationships Stored:** 1 of 4 confirmed

> **Note:** This analysis identifies relationships based on **solving the same problem**, not shared technology.

---

## ✅ Successfully Stored Relationships

### 1. CodeMap ↔ Code-Graph (Frontend-Backend)

**Relationship Type:** Frontend-Backend Pair
**Confidence:** 0.95

**Description:**
Both projects solve the core problem of code structure understanding and visualization for TypeScript/JavaScript developers. CodeMap provides the backend analysis engine with AST parsing, dependency analysis, complexity calculation, and graph construction capabilities. Code-Graph is the VS Code extension that integrates CodeMap as a library to provide IDE-based visualization and interactive navigation.

**Shared Problem:**
Code comprehension for TS/JS developers through structure analysis and visualization

**Technical Connection:**
- Code-Graph explicitly depends on 'codemap-engine' package
- Same target audience (TS/JS developers)
- Same frameworks supported (NestJS, React, Express, Next.js)
- Complementary: backend analysis engine + IDE visualization frontend

---

## ⚠️ Failed to Store (3 relationships)

### 2. OfficeRider Web ↔ Mobile (Shared Problem)

**Relationship Type:** Shared Problem
**Confidence:** 0.95
**Storage Status:** ❌ Failed

**Description:**
Both projects solve the identical core problem: trusted carpool matching for office commuters. They target the same user base (professionals needing daily commute partners), provide the same fundamental features (verification-based trust system, ride posting, ride requests, route matching, in-app messaging), and serve the same job-to-be-done (forming lasting carpool partnerships for recurring commutes).

**Projects:**
- 12e02daa (Next.js web) - More mature with 16 features
- 3f56c016 (Flutter mobile) - Earlier stage with 10 features

**Shared Problem:**
Trusted office carpooling for recurring commutes with verification-based trust system

---

### 3. InvoiceX ↔ Invoice-App (Shared Problem)

**Relationship Type:** Shared Problem
**Confidence:** 0.92
**Storage Status:** ❌ Failed

**Description:**
Both projects solve the core invoice generation problem for freelancers and small businesses. InvoiceX Frontend (Nuxt.js) and invoice-app (SvelteKit) provide invoice creation, real-time editing, professional output generation (PDF vs print-optimized), and user interface for managing invoice line items.

**Projects:**
- d7769b15 (InvoiceX Frontend - Nuxt.js)
- f06de655 (invoice-app - SvelteKit)

**Shared Problem:**
Helping freelancers/small businesses create professional invoices quickly

**Relationship:** Direct competitors with alternative technical implementations

---

### 4. aidnet → aidnet-lite (Evolution)

**Relationship Type:** Evolution
**Confidence:** 0.95
**Storage Status:** Already stored in previous analysis

**Description:**
aidnet (comprehensive humanitarian platform design) evolved into aidnet-lite (production MVP implementation). Both are humanitarian coordination platforms for disaster relief with identical technology stack (Next.js, Supabase, PostgreSQL, TypeScript, Tailwind CSS).

**Projects:**
- bca35bd7 (aidnet) - Conceptual platform (April 2026)
- eb537d62 (aidnet-lite) - Production MVP (July 2026)

**Shared Problem:**
Humanitarian coordination for disaster relief

**Evolution:**
- Comprehensive planning → Focused production implementation
- 55KB PRD → 9.7KB simplified PRD
- Complex features → Core functionality only

---

## 📊 Summary

| Metric | Count |
|--------|-------|
| **Candidates Found** | 9 |
| **Confirmed** | 4 |
| **Successfully Stored** | 1 |
| **Failed to Store** | 3 |

---

## 🎯 Key Insights

### Problem-Domain Clusters

1. **Code Comprehension** - CodeMap + Code-Graph (Frontend-Backend pair)
2. **Trusted Carpooling** - OfficeRider Web + Mobile (Platform pair)
3. **Invoice Generation** - InvoiceX + invoice-app (Competitive implementations)
4. **Humanitarian Coordination** - aidnet + aidnet-lite (Evolution)

### Relationship Types

- **Frontend-Backend Pairs:** 1 (Code comprehension tools)
- **Platform Pairs:** 1 (Web + Mobile carpooling)
- **Competitive Solutions:** 1 (Invoice generation)
- **Evolution:** 1 (aidnet → aidnet-lite)

### Action Items

1. **Fix Storage Issues** - 3 relationships failed to store, investigate database constraints
2. **Consolidate Competitive Projects** - InvoiceX and invoice-app solve same problem
3. **Leverage Platform Synergy** - OfficeRider web + mobile share core business logic
4. **Double Down on Code Tools** - CodeMap + Code-Graph have strong frontend-backend integration

---

## 🔍 Technical Debt

**Storage Failures:**
- 3 of 4 relationships failed to store in database
- Storage agent returned `[true, false, false, false]`
- Need to investigate database constraints or validation errors

**Recommendation:**
Re-run storage phase with error logging to identify root cause of failures.
