# UXGuidelines.md

# Portfolio Dashboard UX Guidelines

Version: 0.1 (Draft)

> This document derives from PlatformSpecification.md dashboard specification and Guideline.md. All dashboard design should follow these guidelines.

---

## Core Principle

**The Dashboard is Read-Only**

The dashboard visualizes knowledge. It must NOT:
- Invoke AI
- Modify repositories
- Perform analysis
- Modify knowledge

All semantic work belongs to AI agents. All modifications belong to the CLI.

---

## Dashboard Pages

Per PlatformSpecification.md, the dashboard consists of five pages:

### 1. Portfolio Overview

**Purpose:** High-level portfolio awareness

**Content:**
- Counts (total projects, active projects)
- Activity (recent modifications, new discoveries)
- Technologies (summary across portfolio)

### 2. Project List

**Purpose:** Browse and filter all projects

**Capabilities:**
- Search
- Filters
- Sorting

### 3. Project Detail

**Purpose:** Comprehensive view of a single project

**Content:**
- Metadata
- Documentation
- Analysis (if available)
- Relationships

### 4. Relationship Explorer

**Purpose:** Explore connections between projects

**Content:**
- Visual representation of relationships
- Navigation between related projects

### 5. Statistics

**Purpose:** Portfolio-wide insights

**Content:**
- Technology distribution
- Project maturity
- Activity timelines

---

## Design Principles

### Knowledge Visualization

The dashboard exists to visualize deterministic knowledge and optional semantic analysis. Design for:
- Quick scanning
- Pattern recognition
- Relationship discovery

### Progressive Enhancement

Projects are useful immediately after deterministic discovery. AI analysis enriches the view later. Design for:
- Base layer: deterministic metadata
- Enhancement layer: semantic analysis (when present)

### Single Knowledge Model

Every dashboard view operates on the same canonical model defined in KnowledgeModel.md:
- Project
- Documentation
- Analysis
- Feature
- Technology
- Relationship

No dashboard page should invent its own representation.

---

## User Experience Patterns

### Search-First Navigation

Users discover projects through search (per PRD.md goal: "Search projects instantly").

Search should be:
- Prominently available
- Fast (local search)
- Rich (search metadata, documentation, technologies)

### Non-Destructive Interaction

No dashboard action modifies data:
- Links navigate
- Filters refine
- Sorting reorders
- Selection reveals detail

### Consistent Entity Views

Every project is represented using the same structure:
- Identity (name, path, type)
- Metadata (git, languages, frameworks)
- Documentation (indexed documents)
- Analysis (optional, if analyzed)

---

## Accessibility

While specific accessibility requirements are not yet defined in the specifications, the dashboard should support:
- Keyboard navigation
- Screen reader compatibility
- Sufficient contrast
- Text scaling

Specific standards (WCAG level) to be determined during Milestone 3 implementation.

---

## Visual Design

### Color Scheme

To be defined during Milestone 3 (Dashboard Frontend) implementation.

### Typography

To be defined during Milestone 3 (Dashboard Frontend) implementation.

### Layout

To be defined during Milestone 3 (Dashboard Frontend) implementation.

---

## Interaction Guidelines

### Dashboard → AI Agent Handoff

When a user needs semantic understanding or actions, the pattern is:
1. User explores in dashboard
2. User asks their AI coding agent
3. Agent uses MCP to access Portfolio
4. Agent provides enriched response

The dashboard does not invoke AI directly.

### Dashboard → CLI Handoff

Administrative tasks (initialization, upgrades, diagnostics) flow to the CLI. The dashboard may provide:
- Status indicators
- Links to relevant commands
- Diagnostic information

But the dashboard itself does not execute administrative actions.

---

## Alignment with Product Principles

Per Guideline.md product philosophy:

> The ideal user journey is: Install → Initialize → Forget

The dashboard supports this by:
- Providing frictionless exploration
- Requiring no manual data entry
- Reflecting current state without user maintenance
- Displaying what already exists

---

## Future Extensions

Per Tasks.md future milestones, potential additions include:
- Timeline views
- Portfolio snapshots
- Team workspaces

These will be specified with UX patterns when reaching those milestones.
