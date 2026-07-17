---
name: travel-planner
description: >-
  Plan, draft, map, and iteratively refine multi-day trips by orchestrating
  real-time inventory from the flyai skill (Fliggy). Use whenever the user
  describes a trip with multiple elements such as transport, lodging, places,
  food, dates, or a budget, even when they do not explicitly say "plan". Start
  with a coordinate-backed candidate map and POI briefing, discuss the rough
  area-by-area plan with the user, then produce and continuously revise a
  detailed self-contained HTML guide with markers, popup cards, bookings,
  budget, and checklist. Do NOT use for one isolated lookup such as "find one
  flight" or "search one hotel"; let flyai handle those directly.
---

# Travel Planner

Build multi-day itineraries as a collaboration, not a one-shot answer. Gather
real inventory with **flyai**, place candidate locations on a map, explain the
choices, agree on the broad shape of the trip, and only then commit to a
detailed timeline. Keep the generated HTML as the living plan throughout the
conversation.

## Prerequisites

This skill depends on flyai for travel inventory and booking data.

1. Install the CLI: `npm i -g @fly-ai/flyai-cli`
2. Install the flyai skill: `npx skills add alibaba-flyai/flyai-skill`
3. Verify: `flyai keyword-search --query "things to do in Sanya"`
4. Optional enhanced results:
   `flyai config set FLYAI_API_KEY "your-key"`

Before planning, run `command -v flyai` and confirm that the flyai skill's
`references/` directory is readable. If either is missing, print the install
commands and stop. Never replace unavailable inventory with invented data.

## Ownership boundaries

| Layer | Owner | Responsibility |
| --- | --- | --- |
| Inventory | flyai | Prices, schedules, availability, booking links, POI/hotel images, and any coordinates returned by the provider |
| Enrichment | travel-planner | Normalize records, resolve missing coordinates, validate images, classify candidates, and preserve source provenance |
| Collaboration | travel-planner + user | Review candidate markers, choose priorities, agree on pace and area allocation, and resolve tradeoffs |
| Synthesis | travel-planner | Sequence confirmed places, attach transport and stays, calculate budget, and maintain the HTML guide |

Do not write a competing flight/hotel/POI search implementation. The planner
decides what to query and how to combine the results; flyai executes inventory
queries.

## Interaction model

Use these stages in order. The two user checkpoints are important because they
prevent a polished but unwanted itinerary from becoming expensive to unwind.

```
requirements
  -> inventory + coordinate/image enrichment
  -> candidate markers + POI briefing
  -> CHECKPOINT 1: user selects priorities
  -> rough area-by-area plan
  -> CHECKPOINT 2: user agrees on shape and pace
  -> initial detailed HTML
  -> discuss -> revise data + HTML -> discuss -> ...
```

If the user explicitly asks for an immediate full plan, still perform marker
and area checks internally, then deliver the initial detailed HTML in the same
turn. Label assumptions and invite targeted revisions instead of blocking.

## Workflow

### 1. Parse the request

Extract:

- destination(s) and origin(s)
- exact dates or date range
- traveler count and type
- total or per-person budget and currency
- pace and interests
- lodging preferences
- preferred transport modes
- fixed commitments, must-visit places, and hard exclusions
- desired output path, when supplied

Resolve relative dates by running `date +%Y-%m-%d`. Ask only about blocking
unknowns such as origin or dates. Choose conservative defaults for everything
else and state them.

### 2. Gather inventory with flyai

Before every flyai command, read its current `references/<cmd>.md`; do not
guess flags or copy parameter tables into this skill.

Use discovery commands for fuzzy requests:

| Command | Use |
| --- | --- |
| `ai-search` | Multi-element exploratory intent |
| `keyword-search` | Cross-category products, events, SIMs, packages, or unusual requests |

Use structured commands when dates and cities are known:

| Command | Use |
| --- | --- |
| `search-flight` | Flights and filters |
| `search-train` | Trains and seat classes |
| `search-hotel` | General hotels |
| `search-poi` | Attractions, food, districts, and activities |
| `search-marriott-hotel` | Marriott portfolio only |
| `search-marriott-package` | Marriott packages and experiences |

Parallelize independent hotel, transport, and POI queries. Search enough POIs
to give the user meaningful choices: include likely scheduled places and a
smaller set of credible alternatives rather than only the first N results.

### 3. Normalize and enrich map records

Convert every candidate location into this shared data contract. The map,
attraction cards, hotel cards, and timeline must all refer to the same `id`.

```js
{
  id: "asakusa-sensoji",
  name: "浅草寺",
  type: "attraction",        // attraction | food | lodging | transport
  status: "candidate",       // candidate | scheduled | optional
  day: null,                 // integer after scheduling
  lat: 35.7148,
  lng: 139.7967,
  address: "东京都台东区浅草2-3-1",
  district: "浅草",
  hours: "06:00-17:00",
  duration: "约1.5小时",
  description: "东京最古老的寺院和仲见世商店街。",
  image: "https://...",
  price: "免费",
  bookingUrl: "https://...",
  source: "flyai",
  coordinateSource: "flyai | nominatim | verified-manual",
  imageSource: "flyai | official | commons"
}
```

Rules:

- IDs are stable slugs. Do not change them when a POI moves to another day.
- `candidate` markers are gray; `scheduled` markers use day colors;
  `optional` markers remain gray/dashed.
- Preserve source fields so later revisions can distinguish verified data from
  editorial notes.
- Never emit `[0,0]`, a city-center placeholder, or guessed coordinates. Omit
  an unresolved marker and list it in the map status instead.

#### Coordinate acquisition

Resolve coordinates during generation, never through browser-side bulk
geocoding.

1. Inspect the actual flyai JSON for coordinate-like fields (`lat`/`lng`,
   `latitude`/`longitude`, or a structured location). Normalize numeric strings
   and validate latitude `[-90,90]` and longitude `[-180,180]`.
2. If coordinates are absent, query a geocoder with the most specific string
   available: `POI name + full address + district + city + country`.
3. For public Nominatim, use
   `https://nominatim.openstreetmap.org/search` with `format=jsonv2`, `limit=3`,
   `addressdetails=1`, and an appropriate `countrycodes` filter. Send an
   identifiable User-Agent, use one thread, stay below one request per second,
   and cache each normalized query. Do not use Nominatim for autocomplete.
4. Compare returned `display_name`, city, district, and country before choosing
   a result. If ambiguous, search the venue's official address or ask the user
   instead of selecting the first match blindly.
5. Record `coordinateSource`. Reuse cached results on later revisions.

Public Nominatim and OSM tile services are community infrastructure. Keep the
provider swappable, honor their current policies, and avoid bulk/offline tile
downloads.

#### Image acquisition

Use this priority order:

1. flyai `picUrl` for POIs or `mainPic` for hotels
2. another flyai/Fliggy result for the exact same entity
3. the venue or official tourism site's representative image
4. a clearly licensed Wikimedia Commons image

Do not reuse a vaguely related photo to fill a blank card. Validate every
chosen URL with an actual request and, for the final HTML, verify that the image
has non-zero rendered dimensions. Store `imageSource`. Render remote images
with `loading="lazy"`, descriptive `alt`, `referrerpolicy="no-referrer"` when
the image host requires it, and a graceful `onerror` fallback that hides the
image area instead of showing a broken icon.

### 4. Marker-first candidate review

Before assigning detailed times:

1. Render or update the template's map using all coordinate-backed candidates.
2. Group the same records into a compact POI catalog by district.
3. Introduce each candidate in one or two useful lines: why visit, typical
   duration, opening constraint, price, and any booking requirement.
4. Separate strong recommendations from optional alternatives. Do not imply
   that every marker will fit.
5. Ask the user which places are must-do, optional, or unwanted, and whether
   the geographic spread matches their expectations.

Use the Question tool when available. Keep the checkpoint decision-oriented;
do not ask the user to re-enter facts already supplied.

### 5. Discuss the rough plan

After checkpoint 1, cluster confirmed POIs geographically and propose the broad
shape of the trip:

- one primary district or connected corridor per day
- hotel/base location and major inbound/outbound legs
- relaxed versus packed alternatives when there is a real tradeoff
- fixed-date constraints, closing days, event times, and bad-weather fallback
- candidates that were omitted and why

Present the proposed day-to-area allocation without pretending exact times are
settled. Ask for confirmation or focused changes. This is checkpoint 2.

### 6. Produce the initial detailed plan

Once the rough plan is accepted:

1. Sequence each day's confirmed POIs to minimize backtracking.
2. Add realistic travel buffers and meal/rest windows.
3. Attach flights/trains at the trip edges and lodging to every night.
4. Check opening hours, closed days, event dates, check-in/out, and departure
   buffers before locking times.
5. Total flyai prices × travelers/nights. Clearly label food, local transit,
   and contingency estimates that do not come from flyai.
6. Surface every available `jumpUrl`/`detailUrl` as a booking link.
7. Render the initial HTML using
   `references/travel-guide-template.html`.

Save as `itinerary-<destination>-<N>d.html`, unless the user supplied a path.

### 7. Continue the planning loop

Treat the first detailed plan as version 1, not the end of the conversation.
For every later user change:

1. identify affected POIs, days, transport, budget, and map layers
2. preserve unrelated user decisions
3. update the shared POI records first
4. regenerate marker colors/status, cards, timeline, route overview, and budget
5. validate the whole HTML
6. summarize what changed and surface the next consequential decision

Overwrite the same HTML path so the user always has one current guide. Create
versioned copies only when explicitly requested.

## HTML template contract

`references/travel-guide-template.html` is canonical. Fill these sections:

| Section | Source |
| --- | --- |
| Hero | title, destination, dates, travelers, budget |
| Weather | current forecast or clearly labeled seasonal guidance |
| Route overview | agreed day-to-area sequence |
| Interactive map | normalized `MAP_POIS` records with valid coordinates |
| Daily itinerary | scheduled records and transport moves |
| Accommodation | hotel records, images, prices, booking links |
| Food | restaurant records and reservation notes |
| Attractions | candidate/scheduled records; cards use `data-poi-id` |
| Budget | flyai totals plus labeled estimates |
| Checklist | destination-specific preparation |
| References | official tourism, transit, venue, map, and booking sources |

### Map behavior

- Use Leaflet 1.9.4 and `https://tile.openstreetmap.org/{z}/{x}/{y}.png`.
- Keep the visible OpenStreetMap contributor attribution; do not hide it.
- Do not prefetch or package tiles for offline use.
- Fit to valid marker bounds with padding and a reasonable `maxZoom`; if there
  is one point, center on it; if there are none, hide the map and show why.
- Do not cluster by default. When markers overlap, clicking one should expand
  the colocated markers radially so each remains selectable.
- Legend controls toggle scheduled days, candidates, hotels, and transport.
- A fullscreen button uses the browser Fullscreen API and calls
  `map.invalidateSize()` after the size transition.
- Clicking a POI/hotel card with `data-poi-id` focuses its marker and opens the
  same popup card. Keyboard Enter/Space must do the same.

### Popup card format

Use one escaped `buildPopup(record)` implementation for every marker. Keep the
field order stable:

1. representative image (when valid)
2. status/day and type badges
3. name
4. address
5. opening hours and suggested duration
6. concise description
7. price and important note
8. booking/detail link

Escape all text and validate URLs before inserting record data into HTML. Open
external links with `target="_blank" rel="noopener noreferrer"`.

## Validation before delivery

- every scheduled POI has one stable ID used by map, card, and timeline
- every marker has finite, in-range coordinates and a recorded source
- unresolved places are reported and omitted, not guessed
- every image URL was requested and final rendered images are nonblank
- popup fields are escaped and external URLs use safe protocols
- OSM attribution is visible
- cards open the matching marker by mouse and keyboard
- exact-overlap expansion, legend toggles, fullscreen, and fitBounds work
- no duplicate IDs, broken labels, JavaScript syntax errors, or console errors
- desktop and mobile screenshots show a nonblank map without overlap
- prices, schedules, availability, and booking links remain traceable to flyai

## Hard rules

- Never fabricate price, schedule, availability, coordinates, opening hours,
  or image provenance.
- Include "基于 fly.ai 实时数据" in the guide.
- Read live flyai command references before each call.
- Parallelize independent inventory calls, but geocode sequentially when the
  provider policy requires it.
- Label estimates explicitly.
- Preserve the user's confirmed decisions across iterations.
- Do not jump from raw search results directly to a polished detailed plan
  unless the user explicitly asks to skip discussion.

## What not to do

- Do not trigger for one isolated flight, train, hotel, or attraction lookup.
- Do not copy flyai parameter tables into this skill.
- Do not geocode repeatedly in the generated browser page.
- Do not use placeholder coordinates or unrelated stock images.
- Do not hide candidates merely because they did not enter the first draft;
  keep useful alternatives as gray optional markers.
- Do not end after writing version 1 when the user is still discussing changes.
