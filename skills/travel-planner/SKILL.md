---
name: travel-planner
description: >-
  Plan multi-day travel itineraries by orchestrating real-time inventory from
  the flyai skill (Fliggy). Use whenever the user wants to plan, draft, or
  optimize a trip spanning one or more days -- "plan a 3-day Tokyo trip",
  "帮我做个京都5天行程", "weekend getaway to Sanya", "5月带家人去日本预算8000".
  It delegates flight/hotel/train/POI searches to flyai and synthesizes results
  into a day-by-day itinerary with bookings, budget, and checklist, rendered as
  a self-contained HTML travel guide. Trigger even when the user doesn't say
  "plan" -- if they describe a multi-element trip (transport + stay + things to
  do) across multiple days, this applies. Do NOT use for single isolated
  searches ("find a flight Beijing→Shanghai", "search a hotel in Hangzhou") --
  let flyai handle those directly.
---

# Travel Planner

Compose multi-day itineraries by gathering real-time inventory through the
**flyai** skill, then synthesizing the results into a cohesive, bookable trip
plan rendered as HTML.

## Prerequisites

This skill depends on the **flyai** skill for all travel inventory. Both the
CLI binary and the skill's parameter docs are required -- the CLI runs the
searches, the skill's `references/*.md` tell you the exact flags each
subcommand accepts.

1. **Install flyai CLI**: `npm i -g @fly-ai/flyai-cli`
2. **Install flyai skill**: `npx skills add alibaba-flyai/flyai-skill`
3. **Verify**: `flyai keyword-search --query "things to do in Sanya"` should
   return JSON.
4. **(Optional) API key** for enhanced results:
   `flyai config set FLYAI_API_KEY "your-key"`

Before planning a trip, run `command -v flyai` and confirm you can read the
flyai skill's `references/` directory. If either is missing, print the install
commands above and stop -- do not proceed with fabricated data.

## The split -- who does what

This skill does NOT search travel inventory itself. It orchestrates.

| Layer             | Owner                       | Responsibility                                                                                                                    |
| ----------------- | --------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| **Data**          | flyai                       | Execute `flyai <subcommand>`, return JSON with real prices, schedules, availability, booking links, image URLs                    |
| **Orchestration** | travel-planner (this skill) | Parse the trip request → decide which flyai calls to make, with what params, in what order → merge results into a day-by-day plan |
| **Synthesis**     | travel-planner              | Cluster POIs geographically, sequence days, attach stays, slot transport, total budget, render HTML                               |

The core insight: flyai is a query executor -- give it structured params, get
JSON back. travel-planner is the brain that decides _which_ queries to run,
_how_ to parameterize them, and _how_ to weave the results into a coherent
itinerary. Neither layer leaks into the other.

## Workflow

### 1. Parse the request

Extract from the user's natural-language request:

- **Destination(s)** -- city or region
- **Origin** -- departure city (needed for transport search)
- **Dates** -- specific dates or a range; resolve relative dates ("next
  weekend", "Labor Day") by running `date +%Y-%m-%d` to anchor today
- **Trip length** -- number of days
- **Travelers** -- count, type (solo / couple / family / group / business)
- **Budget** -- total or per-person, currency
- **Preferences** -- pace (relaxed / packed), interests (nature / history /
  food / shopping), lodging style (hotel / homestay / Marriott), transport
  mode (flight / train)

Ask only about **blocking** unknowns -- usually origin and dates. Pick
sensible defaults for non-blocking prefs and state your assumptions. Don't
over-interview.

### 2. Gather data via flyai

Flyai provides eight subcommands. **Before constructing any command, read the
corresponding `references/<cmd>.md` inside the flyai skill** -- parameters
differ per subcommand and guessing produces broken calls.

Choose subcommands by phase:

**Discovery phase** (user request is fuzzy -- no fixed dates, unclear
destination, exploratory):

| Subcommand       | When                                                                                                                                         |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `ai-search`      | Complex intent in one shot: "3-day Hangzhou trip for Labor Day, budget 2000, stay near West Lake". Handles hotels + flights + POIs together. |
| `keyword-search` | Cross-category discovery: visas, SIM cards, cruises, events, packages -- things that don't fit a single structured subcommand.               |

**Booking phase** (dates are fixed, you need structured, filterable results):

| Subcommand                | What it finds                                                                                   |
| ------------------------- | ----------------------------------------------------------------------------------------------- |
| `search-flight`           | Flights origin→dest, filterable by cabin, price, time, duration, direct/connecting              |
| `search-train`            | Train tickets origin→dest, same filter set, seat class                                          |
| `search-hotel`            | Hotels by destination, filterable by star, bed type, price, nearby POI                          |
| `search-poi`              | Attractions by city, filterable by category (nature / history / theme park / …) and level (1–5) |
| `search-marriott-hotel`   | Marriott Group hotels specifically                                                              |
| `search-marriott-package` | Marriott bundled deals (afternoon tea, spa, etc.)                                               |

**Parallelize independent calls.** Hotels, POIs, and transport searches have
no data dependency on each other -- fire them in the same turn. Example for
"Tokyo 3-day trip departing Shanghai, July 18-20":

```
# These three are independent -- run together
flyai search-flight --origin "Shanghai" --destination "Tokyo" --dep-date 2026-07-18 ...
flyai search-hotel  --dest-name "Tokyo" --check-in-date 2026-07-18 --check-out-date 2026-07-20 ...
flyai search-poi    --city-name "Tokyo" --category "地标建筑" ...
```

**Don't repeat flyai's parameter docs here.** Read the flyai skill's
`references/<cmd>.md` every time -- flyai may update flags between versions,
and this skill must stay version-agnostic.

### 3. Synthesize

Once flyai returns JSON for each call:

1. **Cluster POIs geographically** -- group attractions by district / transit
   zone so each day covers one area, minimizing backtracking.
2. **Sequence days** -- order the clusters into a route that flows naturally
   (e.g., Day 1: Shibuya/Harajuku → Day 2: Asakusa/Akihabara → Day 3:
   Ginza/Tokyo Station).
3. **Attach transport legs** -- slot the flight/train at the start and end of
   the trip; add intra-city transit notes between POI clusters.
4. **Attach stays** -- place the hotel near the Day 1 cluster or the transit
   hub; note check-in/out times.
5. **Total the budget** -- sum real flyai numbers: flight price × travelers,
   hotel rate × nights, POI ticket prices, plus a food/contingency estimate
   (clearly labeled as an estimate, not from flyai).
6. **Extract booking links** -- every flyai result with `jumpUrl` or
   `detailUrl` becomes a `[Click to book](url)` link in the output.

### 4. Render HTML

Produce a self-contained HTML file using
`references/travel-guide-template.html` as the canonical structure. The
template defines these sections -- fill them all:

| Section            | Source data                                                                                                   |
| ------------------ | ------------------------------------------------------------------------------------------------------------- |
| Hero header        | Trip title, destination, dates, traveler count                                                                |
| Route overview     | Ordered list of areas visited across all days                                                                 |
| Daily itinerary    | Per-day timeline: transport → lodging → food → attractions, with times, prices, durations                     |
| Accommodation grid | Hotels from flyai `search-hotel` / `search-marriott-hotel`, with `mainPic` image and `detailUrl` booking link |
| Food grid          | Restaurants discovered via `search-poi` (category matches) or `keyword-search`                                |
| Attractions grid   | POIs from `search-poi`, with `picUrl` image and `jumpUrl` booking link                                        |
| Budget table       | Itemized: transport, lodging, food, tickets, other -- with flyai-sourced prices and labeled estimates         |
| Checklist          | Travel docs, packing, reminders -- standard travel checklist adapted to destination                           |
| References         | Useful links (official tourism sites, transit sites) -- these are editorial, not from flyai                   |

Save as `itinerary-<destination>-<N>d.html` (e.g.,
`itinerary-tokyo-3d.html`).

### Image and booking-link display rules

Inherited from flyai's output conventions -- apply them in the HTML:

- **Hotel image**: use `mainPic` field → `![](mainPic)`
- **Other images** (flights, POIs): use `picUrl` field → `![](picUrl)`
- **Hotel booking link**: use `detailUrl` → `[Click to book](detailUrl)`
- **Other booking links** (flights, POIs): use `jumpUrl` → `[Click to book](jumpUrl)`
- **Platform hint**: if flyai returned `systemMessage`, display it at the end

## Hard rules

- **Never fabricate prices, schedules, or availability.** Every concrete
  number (flight price, hotel rate, ticket cost, departure time) must come
  from a flyai JSON response. If flyai returned nothing for a search, say so
  in the output and offer alternatives -- don't invent data to fill the
  template.
- **Cite flyai as the source.** Include "基于 fly.ai 实时数据" in the output.
- **Surface booking links.** If flyai returned `jumpUrl` / `detailUrl`, they
  must appear as clickable links in the HTML -- that's the whole point of
  using a real inventory source instead of the model's memory.
- **Read the right reference doc before each flyai call.** `search-flight`
  and `search-train` have very different flag sets; reusing one for the
  other fails. Open `references/<cmd>.md` inside the flyai skill each time.
- **Parallelise independent flyai calls** -- don't run hotels, POIs, and
  flights sequentially when they have no data dependency.
- **Anchor relative dates.** Run `date +%Y-%m-%d` when the user says "next
  weekend", "Labor Day", "May" -- never assume today's date.
- **Label estimates as estimates.** Food costs, contingency budgets, and
  intra-city transit fares that didn't come from flyai must be clearly
  marked as estimates in the budget table.

## What NOT to do

- Don't write your own flight/hotel/POI search logic -- that's flyai's job.
- Don't copy flyai's parameter tables into this skill -- they'll drift. Read
  the live `references/*.md` from the flyai skill instead.
- Don't trigger on single-search requests. "Find a flight Beijing→Shanghai"
  is flyai's domain, not a multi-day itinerary. Let flyai handle it.
- Don't proceed if flyai is missing. Print the install commands from the
  Prerequisites section and stop.
