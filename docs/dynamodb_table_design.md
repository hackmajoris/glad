# DynamoDB Test Queries - AWS CLI Commands

This document contains executable AWS CLI commands to test all query patterns for the GLAD entities table.

---

## Table Overview

### Main Table: `glad-entities`

**Table Structure (from CDK):**
- **Partition Key (PK):** `EntityType` (String) - Entity type discriminator
- **Sort Key (SK):** `entity_id` (String) - Unique identifier for each entity
- **Billing Mode:** PAY_PER_REQUEST (on-demand)
- **Point-in-Time Recovery:** Disabled
- **Deletion Protection:** Enabled

**Entity Types:**
- `User` - User profiles
- `Skill` - Master skill definitions (catalog)
- `UserSkill` - User-to-skill relationships with proficiency data

---

## The Power of Multi-Key Composite GSI

**With just ONE Global Secondary Index using composite keys, you can create 15+ different query patterns!**

This is the power of DynamoDB's multi-key GSI design. Instead of creating multiple indexes for different query patterns, you strategically design a single index with composite keys that enables flexible querying at multiple levels of granularity.

### Single GSI: `BySkill`

**Composite Partition Key (1 attribute):**
- `Category` (String) - Skill category for broad partitioning

**Composite Sort Keys (4 attributes):**
1. `SkillName` (String) - Specific skill name
2. `ProficiencyLevel` (String) - Skill proficiency level
3. `YearsOfExperience` (Number) - Years of experience with the skill
4. `Username` (String) - User identifier (ensures uniqueness)

**Why This Design is Powerful:**

✅ **Hierarchical Querying:** Query from broad (Category) to specific (down to individual Username)
✅ **Partial Sort Keys:** Use just SK1, or SK1+SK2, or SK1+SK2+SK3, etc. - DynamoDB allows left-to-right querying
✅ **Range Operations:** Apply `>=`, `<=`, `BETWEEN` on the last sort key in your query
✅ **Efficient Sorting:** Data is automatically sorted by the composite key order
✅ **Single Index Cost:** Pay for only one GSI instead of multiple indexes

**Query Flexibility Examples:**
- Query 1: `Category = "Programming"` → All programming skills
- Query 2: `Category = "Programming" AND SkillName = "Python"` → All Python users
- Query 3: `Category = "Programming" AND SkillName = "Python" AND ProficiencyLevel = "Expert"` → Python experts
- Query 4: `... AND YearsOfExperience >= 5` → Experienced Python experts
- Query 5: `... AND Username = "john"` → Specific user check

**All from ONE index!**

---

## Sample Data Structure

### Main Table Sample Items

| EntityType  | entity_id                   | Additional Attributes                                                                                   | Description                   |
|-------------|-----------------------------|---------------------------------------------------------------------------------------------------------|-------------------------------|
| `User`      | `USER#john_doe`             | Username, Name, Email, CreatedAt, UpdatedAt                                                             | User profile                  |
| `Skill`     | `SKILL#python`              | SkillID, SkillName, Category, Description, Tags                                                         | Master skill catalog          |
| `UserSkill` | `USERSKILL#john_doe#python` | Username, SkillID, SkillName, Category, ProficiencyLevel, YearsOfExperience, Endorsements, LastUsedDate | User's skill with proficiency |

### GSI `BySkill` Sample Items

Here's how UserSkill items appear in the GSI (sorted by composite sort key):

| Category    | SkillName  | ProficiencyLevel | YearsOfExperience | Username      | Endorsements | LastUsedDate |
|-------------|------------|------------------|-------------------|---------------|--------------|--------------|
| Programming | Python     | Beginner         | 1                 | jane_smith    | 5            | 2025-12-10   |
| Programming | Python     | Intermediate     | 3                 | mike_wilson   | 25           | 2025-11-15   |
| Programming | Python     | Advanced         | 5                 | alice_johnson | 65           | 2025-12-15   |
| Programming | Python     | Expert           | 7                 | bob_smith     | 120          | 2025-12-18   |
| Programming | Python     | Expert           | 10                | diana_evans   | 200          | 2025-12-20   |
| Frontend    | TypeScript | Advanced         | 4                 | alex_chen     | 30           | 2025-12-15   |
| Frontend    | TypeScript | Expert           | 6                 | betty_wang    | 75           | 2025-12-18   |
| Backend     | Go         | Expert           | 7                 | alice_smith   | 80           | 2025-12-18   |
| Cloud       | AWS        | Expert           | 9                 | charlie_brown | 150          | 2025-11-20   |
| DevOps      | Docker     | Beginner         | 1                 | tom_davis     | 8            | 2025-10-10   |

**Notice:** Items are naturally sorted by `Category` → `SkillName` → `ProficiencyLevel` → `YearsOfExperience` → `Username`. This sorting enables efficient range queries and pagination.

---

## Key Attributes Explained

### Main Table Keys
- **EntityType:** Discriminator to separate Users, Skills, and UserSkills
- **entity_id:** Hierarchical identifier pattern:
  - `USER#<username>`
  - `SKILL#<skill_id>`
  - `USERSKILL#<username>#<skill_id>`

### GSI Keys (BySkill)
- **Category (PK):** Broad partitioning (Programming, Frontend, Backend, Cloud, DevOps, Database, Mobile, Data, Security, Other)
- **SkillName (SK1):** Specific skill within category
- **ProficiencyLevel (SK2):** Beginner, Intermediate, Advanced, Expert
- **YearsOfExperience (SK3):** NUMBER type for range queries and sorting
- **Username (SK4):** User identifier, ensures uniqueness

### Non-Key Attributes (Available for FilterExpression)
- **Endorsements:** NUMBER - Peer endorsements count
- **LastUsedDate:** STRING (ISO 8601) - When skill was last used
- **Notes:** STRING - Additional skill notes
- **SkillID:** STRING - Immutable skill identifier
- **CreatedAt/UpdatedAt:** STRING (ISO 8601) - Timestamps

---

## GSI Design (Option 1 - Maximum Flexibility)

**Index Name:** `BySkill`

**Partition Key (1 attribute):**
- `Category` (String)

**Sort Keys (4 attributes):**
1. `SkillName` (String)
2. `ProficiencyLevel` (String)
3. `YearsOfExperience` (Number)
4. `Username` (String)

**Note:** Endorsements (Number) and LastUsedDate (String) are not in the GSI keys, but can be used with FilterExpression.

---

## Access Patterns Summary

### Main Table Access Patterns

| # | Pattern Name              | Table/Index | Key Condition                                                                  | Use Case                          | API Endpoint                               |
|---|---------------------------|-------------|--------------------------------------------------------------------------------|-----------------------------------|--------------------------------------------|
| 1 | Get All Users             | Main Table  | `EntityType = "User"`                                                          | List all users in system          | `GET /users`                               |
| 2 | Get All Master Skills     | Main Table  | `EntityType = "Skill"`                                                         | List all master skill definitions | `GET /master-skills`                       |
| 3 | Get All User Skills       | Main Table  | `EntityType = "UserSkill"`                                                     | List all user skill records       | -                                          |
| 4 | Get Specific User         | Main Table  | `EntityType = "User" AND entity_id = "USER#<username>"`                        | Get user profile by username      | `GET /users/{username}`                    |
| 5 | Get Specific Master Skill | Main Table  | `EntityType = "Skill" AND entity_id = "SKILL#<skillID>"`                       | Get master skill details          | `GET /master-skills/{skillID}`             |
| 6 | Get Specific User Skill   | Main Table  | `EntityType = "UserSkill" AND entity_id = "USERSKILL#<username>#<skillID>"`    | Get user's specific skill         | `GET /users/{username}/skills/{skillName}` |
| 7 | Get All Skills for User   | Main Table  | `EntityType = "UserSkill" AND begins_with(entity_id, "USERSKILL#<username>#")` | List all skills for a user        | `GET /users/{username}/skills`             |

### GSI Access Patterns (BySkill Index)

| #  | Pattern Name                  | PK                | SK Condition                                                                                   | Filter                                           | Use Case                                                            | API Endpoint                                                                                 |
|----|-------------------------------|-------------------|------------------------------------------------------------------------------------------------|--------------------------------------------------|---------------------------------------------------------------------|----------------------------------------------------------------------------------------------|
| 8  | All Skills in Category        | `Category = :cat` | -                                                                                              | -                                                | All skills in Programming category                                  | `GET /skills?category=Programming`                                                           |
| 9  | All Users with Skill          | `Category = :cat` | `SkillName = :skill`                                                                           | -                                                | Everyone who knows Python                                           | `GET /skills/python/users`                                                                   |
| 10 | Users at Skill Level          | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level`                                             | -                                                | All Python experts                                                  | `GET /skills/python/users?level=Expert`                                                      |
| 11 | Users with Min Experience     | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level AND YearsOfExperience >= :years`             | -                                                | Python experts with 5+ years                                        | `GET /skills/python/users?level=Expert&minYears=5`                                           |
| 12 | Users by Experience (Desc)    | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level`                                             | `--scan-index-forward false`                     | Most experienced Python experts                                     | `GET /skills/python/users?level=Expert&sort=experience_desc`                                 |
| 13 | Experience Range              | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level AND YearsOfExperience BETWEEN :min AND :max` | -                                                | Intermediate Python devs (2-5 years)                                | `GET /skills/python/users?level=Intermediate&minYears=2&maxYears=5`                          |
| 14 | Check User Skill Level        | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level AND Username = :user`                        | -                                                | Check if john_doe is Python expert                                  | `GET /users/john_doe/skills/python?checkLevel=Expert`                                        |
| 15 | Filter by Endorsements        | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level`                                             | `Endorsements >= :min`                           | Python experts with 50+ endorsements                                | `GET /skills/python/users?level=Expert&minEndorsements=50`                                   |
| 16 | Filter by Last Used           | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level`                                             | `LastUsedDate >= :date`                          | Python experts active in last 6 months                              | `GET /skills/python/users?level=Expert&activeSince=2024-06-20`                               |
| 17 | Multi-Criteria Filter         | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level AND YearsOfExperience >= :years`             | `Endorsements >= :min AND LastUsedDate >= :date` | Senior Python experts (5+ years, 50+ endorsements, recently active) | `GET /skills/python/users?level=Expert&minYears=5&minEndorsements=50&activeSince=2024-06-20` |
| 18 | All Expert Skills in Category | `Category = :cat` | -                                                                                              | `ProficiencyLevel = :level`                      | All expert-level skills in Programming                              | `GET /skills?category=Programming&level=Expert`                                              |
| 19 | Skills with Prefix            | `Category = :cat` | `begins_with(SkillName, :prefix)`                                                              | -                                                | All Java-related skills (Java, JavaScript)                          | `GET /skills?category=Programming&skillPrefix=Java`                                          |
| 20 | Proficiency Distribution      | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level`                                             | `--select COUNT`                                 | Count Python users by proficiency level                             | `GET /skills/python/distribution`                                                            |
| 21 | Top N by Experience           | `Category = :cat` | `SkillName = :skill AND ProficiencyLevel = :level`                                             | `--scan-index-forward false --limit N`           | Top 10 most experienced Go experts                                  | `GET /skills/go/users/top?limit=10&level=Expert`                                             |
| 22 | Users Above Min Level         | `Category = :cat` | `SkillName = :skill`                                                                           | `ProficiencyLevel IN (:level1, :level2)`         | TypeScript users at Advanced or Expert                              | `GET /skills/typescript/users?minLevel=Advanced`                                             |
| 23 | Pagination                    | `Category = :cat` | -                                                                                              | `--limit N --exclusive-start-key`                | Paginated results for category                                      | `GET /skills?category=Programming&page=2&limit=20`                                           |

**Total Access Patterns:** 23 (7 Main Table + 16 GSI)

---

## Prerequisites

```bash
# AWS Profile
export AWS_PROFILE=passbrains-ilisa-amplify

# Table and Index names
TABLE_NAME="glad-entities"
GSI_NAME="BySkill"
```

---

## Main Table Query Patterns

### Pattern 1: Get All Users

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType" \
  --expression-attribute-values '{
    ":entityType": {"S": "User"}
  }'
```

---

### Pattern 2: Get All Master Skills

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType" \
  --expression-attribute-values '{
    ":entityType": {"S": "Skill"}
  }'
```

---

### Pattern 3: Get All User Skills

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType" \
  --expression-attribute-values '{
    ":entityType": {"S": "UserSkill"}
  }'
```

---

### Pattern 4: Get Specific User

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType AND entity_id = :entityId" \
  --expression-attribute-values '{
    ":entityType": {"S": "User"},
    ":entityId": {"S": "USER#john_doe"}
  }'
```

---

### Pattern 5: Get Specific Master Skill

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType AND entity_id = :entityId" \
  --expression-attribute-values '{
    ":entityType": {"S": "Skill"},
    ":entityId": {"S": "SKILL#python"}
  }'
```

---

### Pattern 6: Get Specific User Skill

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType AND entity_id = :entityId" \
  --expression-attribute-values '{
    ":entityType": {"S": "UserSkill"},
    ":entityId": {"S": "USERSKILL#jane_doe#python"}
  }'
```

---

### Pattern 7: Get All Skills for a User

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType AND begins_with(entity_id, :prefix)" \
  --expression-attribute-values '{
    ":entityType": {"S": "UserSkill"},
    ":prefix": {"S": "USERSKILL#john_doe#"}
  }'
```

---

## GSI Query Patterns (BySkill) - Maximum Flexibility!

### GSI Pattern 1: All Skills in a Category

**Use Case:** List all user skills in a specific category (e.g., all Programming skills)
             
```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category" \
  --expression-attribute-values '{
    ":category": {"S": "Programming"}
  }'
```

**Returns:** All UserSkill items in the Programming category

---

### GSI Pattern 2: All Users with a Specific Skill (Any Level)

**Use Case:** Find everyone who has Python, regardless of proficiency level
                        
//todo
```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName" \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":skillName": {"S": "JavaScript"}
  }'
```

**Returns:** All UserSkill items for Python (all proficiency levels)

---

### GSI Pattern 3: Users with Skill at Specific Proficiency Level

**Use Case:** Find all Python experts

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --expression-attribute-values '{
    ":category": {"S": "DevOps"},
    ":skillName": {"S": "Docker"},
    ":level": {"S": "Beginner"}
  }'
```

**Returns:** All Python experts

---

### GSI Pattern 4: Users with Minimum Years of Experience

**Use Case:** Find Python experts with at least 5 years of experience

**Option A: Using Sort Key Condition (Efficient)**

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level AND YearsOfExperience >= :minYears" \
  --expression-attribute-values '{
    ":category": {"S": "DevOps"},
    ":skillName": {"S": "Docker"},
    ":level": {"S": "Beginner"},
    ":minYears": {"N": "1"}
  }'
```

**Returns:** Python experts with 5+ years of experience, sorted by experience (ascending)

---

### GSI Pattern 5: Users Sorted by Experience (Descending)

**Use Case:** Find Python experts ordered by most experienced first

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --no-scan-index-forward \
  --expression-attribute-values '{
    ":category": {"S": "DevOps"},
    ":skillName": {"S": "Docker"},
    ":level": {"S": "Beginner"}
  }'
```

**Returns:** Python experts sorted by experience (most experienced first)

---

### GSI Pattern 6: Experience Range Query

**Use Case:** Find intermediate Python developers with 2-5 years of experience

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level AND YearsOfExperience BETWEEN :minYears AND :maxYears" \
  --expression-attribute-values '{
    ":category": {"S": "DevOps"},
    ":skillName": {"S": "Docker"},
    ":level": {"S": "Beginner"},
    ":minYears": {"N": "1"},
    ":maxYears": {"N": "10"}
  }'
```

**Returns:** Intermediate Python developers with 2-5 years experience

---

### GSI Pattern 7: Check if Specific User Has Skill at Level

**Use Case:** Check if user "john_doe" is a Python expert

```bash
aws dynamodb query \
    --table-name glad-entities --profile passbrains-ilisa-amplify \
    --key-condition-expression "EntityType = :entityType AND entity_id = :entityId" \
    --expression-attribute-values '{
      ":entityType": {"S": "UserSkill"},
      ":entityId": {"S": "USERSKILL#jane_doe#python"}
    }'

```

**Returns:** Single UserSkill item if exists, empty if not

---

### GSI Pattern 8: Filter by Endorsements (Using FilterExpression)

**Use Case:** Find Python experts with at least 50 endorsements

**Note:** Endorsements is NOT in the GSI keys, so we use FilterExpression

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --filter-expression "Endorsements >= :minEndorsements" \
  --expression-attribute-values '{
    ":category": {"S": "DevOps"},
    ":skillName": {"S": "Docker"},
    ":level": {"S": "Beginner"},
    ":minEndorsements": {"N": "2"}
  }'
```

**Returns:** Python experts with 50+ endorsements (filtered after query)

---

### GSI Pattern 9: Filter by Last Used Date (Using FilterExpression)

**Use Case:** Find Beginners who used Docker in the last 6 months

**Note:** LastUsedDate is NOT in the GSI keys, so we use FilterExpression

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --filter-expression "LastUsedDate >= :recentDate" \
  --expression-attribute-values '{
    ":category": {"S": "DevOps"},
    ":skillName": {"S": "Docker"},
    ":level": {"S": "Beginner"},
    ":recentDate": {"S": "2024-06-20"}
  }'
```

**Returns:** Beginners who used Docker since 2024-06-20

---

### GSI Pattern 10: Complex Multi-Criteria Filter

**Use Case:** Find senior Python experts (5+ years, 50+ endorsements, used recently)

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level AND YearsOfExperience >= :minYears" \
  --filter-expression "Endorsements >= :minEndorsements AND LastUsedDate >= :minDate" \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":skillName": {"S": "Python 3"},
    ":level": {"S": "Expert"},
    ":minYears": {"N": "5"},
    ":minEndorsements": {"N": "50"},
    ":minDate": {"S": "2024-06-20"}
  }'
```

**Returns:** Highly qualified, active Python experts

---

### GSI Pattern 11: All Skills at Specific Proficiency (Across All Skills)

**Use Case:** Find all Expert-level skills in Programming category

**⚠️ WARNING:** This query will FAIL because `ProficiencyLevel` is a GSI key attribute (SK2) and cannot be used in FilterExpression.

**Error:** `Filter Expression can only contain non-primary key attributes: Primary key attribute: ProficiencyLevel`

**Workaround:** Query each proficiency level separately or fetch all skills in the category and filter in application code.

```bash
# This query will return an ERROR - kept for reference
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category" \
  --filter-expression "ProficiencyLevel = :level" \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":level": {"S": "Expert"}
  }'
```

**Alternative - Fetch all and filter in code:**
```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category" \
  --expression-attribute-values '{
    ":category": {"S": "Programming"}
  }'
# Then filter ProficiencyLevel = "Expert" in your application code
```

---

### GSI Pattern 12: Skills Starting with Prefix

**Use Case:** Find all users with skills starting with "Java" (Java, JavaScript, etc.)

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND begins_with(SkillName, :skillPrefix)" \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":skillPrefix": {"S": "Java"}
  }'
```

**Returns:** All UserSkills with skills starting with "Java"

---

### GSI Pattern 13: Proficiency Distribution for a Skill (Count by Level)

**Use Case:** Get count of Python users at each proficiency level

**Query 1 - Beginner:**

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --select COUNT \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":skillName": {"S": "Python"},
    ":level": {"S": "Beginner"}
  }'
```

**Query 2 - Intermediate:**

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --select COUNT \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":skillName": {"S": "Python"},
    ":level": {"S": "Intermediate"}
  }'
```

**Query 3 - Advanced:**

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --select COUNT \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":skillName": {"S": "Python"},
    ":level": {"S": "Advanced"}
  }'
```

**Query 4 - Expert:**

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --select COUNT \
  --expression-attribute-values '{
    ":category": {"S": "DevOps"},
    ":skillName": {"S": "Docker"},
    ":level": {"S": "Beginner"}
  }'
```

**Returns:** Count for each proficiency level (4 queries total)

---

### GSI Pattern 14: Top N Users by Experience

**Use Case:** Get top 10 most experienced Go experts

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --limit 10 \
  --expression-attribute-values '{
    ":category": {"S": "Backend"},
    ":skillName": {"S": "Go"},
    ":level": {"S": "Expert"}
  }'
```

**Returns:** Top 10 Go experts by experience

---

### GSI Pattern 15: Users Above Minimum Level (Multiple Levels)

**Use Case:** Find TypeScript users at Advanced OR Expert level

**Note:** This pattern returns ALL proficiency levels, so filter in application code or run separate queries for each level.

```bash
  aws dynamodb query \
    --table-name glad-entities --profile passbrains-ilisa-amplify \
    --index-name BySkill \
    --key-condition-expression "Category = :category AND SkillName = :skillName" \
    --expression-attribute-values '{
      ":category": {"S": "Frontend"},
      ":skillName": {"S": "TypeScript"}
    }'

# Then filter for ProficiencyLevel IN ("Advanced", "Expert") in application code
```

**Alternative - Run two separate queries:**
```bash
# Query 1: Advanced
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --expression-attribute-values '{
    ":category": {"S": "Frontend"},
    ":skillName": {"S": "TypeScript"},
    ":level": {"S": "Advanced"}
  }'

# Query 2: Expert
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName AND ProficiencyLevel = :level" \
  --expression-attribute-values '{
    ":category": {"S": "Frontend"},
    ":skillName": {"S": "TypeScript"},
    ":level": {"S": "Expert"}
  }'
# Merge results in application code
```

**Returns:** TypeScript users at Advanced or Expert level

---

### GSI Pattern 16: Pagination Example

**Use Case:** Get first 20 skills in Programming category, then get next page

**Page 1:**

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category" \
  --limit 20 \
  --expression-attribute-values '{
    ":category": {"S": "Programming"}
  }'
```

**Page 2 (use LastEvaluatedKey from previous response):**

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category" \
  --limit 20 \
  --exclusive-start-key '{"Category":{"S":"Programming"},"SkillName":{"S":"Python"},"ProficiencyLevel":{"S":"Expert"},"YearsOfExperience":{"N":"5"},"Username":{"S":"alice"},"EntityType":{"S":"UserSkill"},"entity_id":{"S":"USERSKILL#alice#python"}}' \
  --expression-attribute-values '{
    ":category": {"S": "Programming"}
  }'
```

---

## Utility Commands

### Count Total Users

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key-condition-expression "EntityType = :entityType" \
  --select COUNT \
  --expression-attribute-values '{
    ":entityType": {"S": "User"}
  }'
```

---

### Count Users with Specific Skill

```bash
aws dynamodb query \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --index-name BySkill \
  --key-condition-expression "Category = :category AND SkillName = :skillName" \
  --select COUNT \
  --expression-attribute-values '{
    ":category": {"S": "Programming"},
    ":skillName": {"S": "Python"}
  }'
```

---

### Get Item by Primary Key (GetItem - faster than Query)

```bash
aws dynamodb get-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --key '{
    "EntityType": {"S": "User"},
    "entity_id": {"S": "USER#john_doe"}
  }'
```

---

## Sample Data for Testing

### Create Test Users

```bash
aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "User"},
    "entity_id": {"S": "USER#john_doe"},
    "Username": {"S": "john_doe"},
    "Name": {"S": "John Doe"},
    "Email": {"S": "john@example.com"},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-01-01T00:00:00Z"}
  }'

aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "User"},
    "entity_id": {"S": "USER#jane_doe"},
    "Username": {"S": "jane_doe"},
    "Name": {"S": "Jane Doe"},
    "Email": {"S": "jane@example.com"},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-01-01T00:00:00Z"}
  }'

aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "User"},
    "entity_id": {"S": "USER#alice_smith"},
    "Username": {"S": "alice_smith"},
    "Name": {"S": "Alice Smith"},
    "Email": {"S": "alice@example.com"},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-01-01T00:00:00Z"}
  }'
```

---

### Create Test Master Skills

```bash
aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "Skill"},
    "entity_id": {"S": "SKILL#python"},
    "SkillID": {"S": "python"},
    "SkillName": {"S": "Python"},
    "Category": {"S": "Programming"},
    "Description": {"S": "Python programming language"},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-01-01T00:00:00Z"}
  }'

aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "Skill"},
    "entity_id": {"S": "SKILL#go"},
    "SkillID": {"S": "go"},
    "SkillName": {"S": "Go"},
    "Category": {"S": "Backend"},
    "Description": {"S": "Go programming language"},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-01-01T00:00:00Z"}
  }'

aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "Skill"},
    "entity_id": {"S": "SKILL#react"},
    "SkillID": {"S": "react"},
    "SkillName": {"S": "React"},
    "Category": {"S": "Frontend"},
    "Description": {"S": "React JavaScript library"},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-01-01T00:00:00Z"}
  }'
```

---

### Create Test UserSkills

```bash
# John - Python Expert (10 years, 150 endorsements)
aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "UserSkill"},
    "entity_id": {"S": "USERSKILL#john_doe#python"},
    "Username": {"S": "john_doe"},
    "SkillID": {"S": "python"},
    "SkillName": {"S": "Python"},
    "Category": {"S": "Programming"},
    "ProficiencyLevel": {"S": "Expert"},
    "YearsOfExperience": {"N": "10"},
    "Endorsements": {"N": "150"},
    "LastUsedDate": {"S": "2025-12-15"},
    "Notes": {"S": "Experienced Python developer"},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-12-15T00:00:00Z"}
  }'

# Jane - Python Advanced (5 years, 42 endorsements)
aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "UserSkill"},
    "entity_id": {"S": "USERSKILL#jane_doe#python"},
    "Username": {"S": "jane_doe"},
    "SkillID": {"S": "python"},
    "SkillName": {"S": "Python"},
    "Category": {"S": "Programming"},
    "ProficiencyLevel": {"S": "Advanced"},
    "YearsOfExperience": {"N": "5"},
    "Endorsements": {"N": "42"},
    "LastUsedDate": {"S": "2025-11-20"},
    "Notes": {"S": ""},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-11-20T00:00:00Z"}
  }'

# Alice - Go Expert (7 years, 80 endorsements)
aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "UserSkill"},
    "entity_id": {"S": "USERSKILL#alice_smith#go"},
    "Username": {"S": "alice_smith"},
    "SkillID": {"S": "go"},
    "SkillName": {"S": "Go"},
    "Category": {"S": "Backend"},
    "ProficiencyLevel": {"S": "Expert"},
    "YearsOfExperience": {"N": "7"},
    "Endorsements": {"N": "80"},
    "LastUsedDate": {"S": "2025-12-18"},
    "Notes": {"S": ""},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-12-18T00:00:00Z"}
  }'

# John - React Intermediate (3 years, 25 endorsements)
aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "UserSkill"},
    "entity_id": {"S": "USERSKILL#john_doe#react"},
    "Username": {"S": "john_doe"},
    "SkillID": {"S": "react"},
    "SkillName": {"S": "React"},
    "Category": {"S": "Frontend"},
    "ProficiencyLevel": {"S": "Intermediate"},
    "YearsOfExperience": {"N": "3"},
    "Endorsements": {"N": "25"},
    "LastUsedDate": {"S": "2025-12-10"},
    "Notes": {"S": ""},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-12-10T00:00:00Z"}
  }'

# Jane - Python Beginner (1 year, 5 endorsements)
aws dynamodb put-item \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --item '{
    "EntityType": {"S": "UserSkill"},
    "entity_id": {"S": "USERSKILL#jane_doe#go"},
    "Username": {"S": "jane_doe"},
    "SkillID": {"S": "go"},
    "SkillName": {"S": "Go"},
    "Category": {"S": "Backend"},
    "ProficiencyLevel": {"S": "Beginner"},
    "YearsOfExperience": {"N": "1"},
    "Endorsements": {"N": "5"},
    "LastUsedDate": {"S": "2025-10-15"},
    "Notes": {"S": ""},
    "CreatedAt": {"S": "2025-01-01T00:00:00Z"},
    "UpdatedAt": {"S": "2025-10-15T00:00:00Z"}
  }'
```

---

## Troubleshooting

### Check if GSI Exists

```bash
aws dynamodb describe-table \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --query "Table.GlobalSecondaryIndexes[?IndexName=='BySkill']"
```

---

### Check GSI Status

```bash
aws dynamodb describe-table \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --query "Table.GlobalSecondaryIndexes[?IndexName=='BySkill'].IndexStatus"
```

---

### List All Attributes in Table

```bash
aws dynamodb describe-table \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --query "Table.AttributeDefinitions"
```

---

### Scan Entire Table (Use Sparingly!)

```bash
aws dynamodb scan \
  --table-name glad-entities --profile passbrains-ilisa-amplify \
  --limit 10
```

---

## Query Pattern Summary

| Pattern | PK       | SK1       | SK2              | SK3                       | SK4      | Use Case               |
|---------|----------|-----------|------------------|---------------------------|----------|------------------------|
| 1       | Category | -         | -                | -                         | -        | All skills in category |
| 2       | Category | SkillName | -                | -                         | -        | All users with skill   |
| 3       | Category | SkillName | ProficiencyLevel | -                         | -        | Users at skill level   |
| 4       | Category | SkillName | ProficiencyLevel | YearsOfExperience >=      | -        | Min experience filter  |
| 5       | Category | SkillName | ProficiencyLevel | YearsOfExperience (desc)  | -        | Most experienced first |
| 6       | Category | SkillName | ProficiencyLevel | YearsOfExperience BETWEEN | -        | Experience range       |
| 7       | Category | SkillName | ProficiencyLevel | YearsOfExperience         | Username | Specific user check    |

---

## Important Notes

1. **Data Types:**
   - `YearsOfExperience`: NUMBER type (e.g., {"N": "5"})
   - `Endorsements`: NUMBER type (e.g., {"N": "42"})
   - No zero-padding needed since both are stored as numbers

2. **Sort Key Hierarchy:**
   - Must query from left to right
   - Can use partial sort keys (SK1, SK1+SK2, SK1+SK2+SK3, etc.)
   - Cannot skip sort keys (e.g., cannot query SK1+SK3 without SK2)

3. **Range Queries:**
   - Only on the LAST sort key specified
   - Operators: `<`, `>`, `<=`, `>=`, `BETWEEN`, `begins_with`

4. **FilterExpression:**
   - For attributes NOT in GSI keys (Endorsements, LastUsedDate)
   - Applied AFTER query (less efficient than KeyConditionExpression)

5. **Sorting:**
   - Default: Ascending (`--scan-index-forward true`)
   - Descending: `--scan-index-forward false`

6. **Pagination:**
   - Use `--limit` to control page size
   - Use `--exclusive-start-key` with LastEvaluatedKey for next page