# DBpedia

## 📚 What is DBpedia?

**DBpedia** is a crowd-sourced community effort to extract structured information from Wikipedia and make it available as a knowledge graph on the web.

### Key Facts About DBpedia

- **Interconnected**: Links to other datasets like Wikidata, GeoNames, and more
- **Free & Open**: Available under open licenses (CC BY-SA and GFDL)
- **Real-world Use**: Powers semantic search, chatbots, recommendation systems, and AI applications

## 🔍 SPARQL

### Understanding SPARQL Basics

#### The Triple Pattern

All data in DBpedia is stored as **triples** in the format:
```
Subject → Predicate → Object
```

**Example:**
```
Albert Einstein → birthDate → 1879-03-14
Albert Einstein → nationality → Germany
```

In SPARQL, this becomes:
```sparql
<http://dbpedia.org/resource/Albert_Einstein> <http://dbpedia.org/ontology/birthDate> "1879-03-14" .
```

#### Basic Query Structure

```sparql
SELECT ?variable1 ?variable2    # What you want to retrieve
WHERE {
  ?subject ?predicate ?object .  # Pattern to match
}
LIMIT 10                         # Optional: limit results
```

### Key SPARQL Concepts

#### 1. Variables
Variables start with `?` and act as placeholders for values you want to find:
- `?person` - will match any person
- `?name` - will match any name
- `?birthDate` - will match any birth date

#### 2. URIs (Resources)
Fixed resources are enclosed in angle brackets:
- `<http://dbpedia.org/resource/Albert_Einstein>` - specific person
- `<http://dbpedia.org/ontology/Scientist>` - type of entity
- `<http://xmlns.com/foaf/0.1/name>` - property for names

#### 3. Common Prefixes
To make queries shorter, use prefixes:
```sparql
PREFIX dbo: <http://dbpedia.org/ontology/>
PREFIX dbr: <http://dbpedia.org/resource/>
PREFIX foaf: <http://xmlns.com/foaf/0.1/>

SELECT ?name WHERE {
  dbr:Albert_Einstein foaf:name ?name .
}
```

### Common SPARQL Query Patterns

#### Pattern 1: Get All Properties of a Resource
**Use Case:** Explore what information is available about something

```sparql
SELECT ?property ?value WHERE {
  <http://dbpedia.org/resource/Albert_Einstein> ?property ?value .
} LIMIT 20
```

**How it works:** The `?` variables will be filled with all properties and their values.

#### Pattern 2: Find Resources by Type
**Use Case:** List all instances of a category (scientists, countries, etc.)

```sparql
SELECT ?scientist ?name WHERE {
  ?scientist a <http://dbpedia.org/ontology/Scientist> .
  ?scientist <http://xmlns.com/foaf/0.1/name> ?name .
} LIMIT 10
```

**Note:** `a` is shorthand for `rdf:type` (means "is a type of")

#### Pattern 3: Filter Results
**Use Case:** Find resources that meet specific criteria

```sparql
SELECT ?country ?name ?population WHERE {
  ?country a <http://dbpedia.org/ontology/Country> .
  ?country <http://xmlns.com/foaf/0.1/name> ?name .
  ?country <http://dbpedia.org/ontology/populationTotal> ?population .
  FILTER (?population > 100000000)
  FILTER (lang(?name) = 'en')
} LIMIT 10
```

**Common Filters:**
- `FILTER (?value > 100)` - numeric comparison
- `FILTER (lang(?name) = 'en')` - language filter
- `FILTER (CONTAINS(?name, "United"))` - text search
- `FILTER (YEAR(?date) = 2020)` - date operations

#### Pattern 4: OPTIONAL Data
**Use Case:** Get data even if some fields are missing

```sparql
SELECT ?person ?name ?birthDate ?deathDate WHERE {
  ?person a <http://dbpedia.org/ontology/Scientist> .
  ?person <http://xmlns.com/foaf/0.1/name> ?name .
  OPTIONAL { ?person <http://dbpedia.org/ontology/birthDate> ?birthDate . }
  OPTIONAL { ?person <http://dbpedia.org/ontology/deathDate> ?deathDate . }
} LIMIT 10
```

**Result:** Returns scientists even if birth/death dates are unknown.

#### Pattern 5: Complex Relationships
**Use Case:** Follow multiple relationships (e.g., find cities in countries)

```sparql
SELECT ?country ?countryName ?city ?cityName WHERE {
  ?country a <http://dbpedia.org/ontology/Country> .
  ?country <http://xmlns.com/foaf/0.1/name> ?countryName .
  ?city <http://dbpedia.org/ontology/country> ?country .
  ?city <http://xmlns.com/foaf/0.1/name> ?cityName .
  FILTER (lang(?countryName) = 'en' && lang(?cityName) = 'en')
} LIMIT 20
```

### Common DBpedia Properties

| Property | URI | Description |
|----------|-----|-------------|
| name | `foaf:name` | Name of the resource |
| abstract | `dbo:abstract` | Short description/summary |
| birthDate | `dbo:birthDate` | Birth date (for people) |
| deathDate | `dbo:deathDate` | Death date (for people) |
| country | `dbo:country` | Associated country |
| capital | `dbo:capital` | Capital city (for countries) |
| population | `dbo:populationTotal` | Population count |
| thumbnail | `dbo:thumbnail` | Image URL |

**Full prefix meanings:**
- `dbo:` = `http://dbpedia.org/ontology/`
- `dbr:` = `http://dbpedia.org/resource/`
- `foaf:` = `http://xmlns.com/foaf/0.1/`

### Should do when use LINKED DATA
- Always use `LIMIT` to prevent overwhelming results
- Filter by language for text properties: `FILTER (lang(?name) = 'en')`
- Use specific URIs when you know the exact resource
- Start simple and add complexity gradually

## SOME APIs THIS PROJECT PROVIDE 

### 1. Execute SPARQL Query
**POST** `/api/v1/sparql` or `/sparql`

Execute a SPARQL query against DBpedia.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/sparql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT ?name WHERE { <http://dbpedia.org/resource/Albert_Einstein> <http://xmlns.com/foaf/0.1/name> ?name . } LIMIT 1"
  }'
```

**Response:**
```json
{
  "head": {
    "vars": ["name"]
  },
  "results": {
    "bindings": [
      {
        "name": {
          "type": "literal",
          "value": "Albert Einstein",
          "xml:lang": "en"
        }
      }
    ]
  }
}
```

---

### 2. Validate SPARQL Query
**POST** `/api/v1/validate`

Validate a SPARQL query without executing it.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/validate \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT ?s ?p ?o WHERE { ?s ?p ?o . } LIMIT 10"
  }'
```

**Response:**
```json
{
  "valid": true,
  "message": "Query is valid"
}
```

## DBPEDIA REPO

https://github.com/aniiyengar/dbpedia

https://github.com/garfix/nli-go

# THIS CURRENT PROJECT'S STRUCTURE

1. **main.go**: Application entry point
   - Loads configuration
   - Initializes DBpedia client
   - Sets up HTTP server
   - **Don't modify** unless changing app initialization

2. **client/dbpedia_client.go**: DBpedia SPARQL client
   - Call `NewDBpediaClient()` to create a client instance
   - Call `Query()` method to execute SPARQL queries
   - **Modify** if you need to change how queries are sent to DBpedia

3. **handlers/handlers.go**: HTTP request handlers
   - Contains `ExecuteSPARQLQuery`, `ValidateSPARQLQuery`, `GetExampleQueries`
   - **Modify** to add new endpoints or change request/response handling

4. **server/server.go**: HTTP server configuration
   - Call `SetupRoutes()` to define API endpoints
   - **Modify** to add new routes or middleware

5. **types/types.go**: Data structures
   - Defines request/response types
   - **Modify** when adding new API endpoints or changing data structures