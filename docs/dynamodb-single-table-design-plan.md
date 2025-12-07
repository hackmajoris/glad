# DynamoDB Single Table Design - Implementation Plan

## Executive Summary

This document outlines the migration from a simple users table to a comprehensive single table design that supports multiple entity types while leveraging DynamoDB's new multi-key composite GSI capabilities (up to 4 partition keys + 4 sort keys).

## Current State Analysis

### Existing Table Structure
```
Table: users
- Partition Key: username (String)
- Sort Key: None
- Attributes: name, password_hash, created_at, updated_at
```

### Current Access Patterns
| Pattern # | Description                  | Type  | Implementation          | RPS Estimate     |
|-----------|------------------------------|-------|-------------------------|------------------|
| AP1       | Get user by username (login) | Read  | GetItem(username)       | High (500/sec)   |
| AP2       | Create new user (register)   | Write | PutItem with condition  | Medium (50/sec)  |
| AP3       | Update user profile          | Write | UpdateItem(username)    | Low (10/sec)     |
| AP4       | Check if user exists         | Read  | GetItem with projection | Medium (100/sec) |
| AP5       | List all users               | Read  | Scan (anti-pattern!)    | Low (1/sec)      |

### Pain Points
1. **No sort key** - Limited query flexibility
2. **Single entity type** - Can't scale to multiple entities
3. **Scan operation** - ListUsers is inefficient and expensive
4. **No relationships** - Can't model related entities

## Proposed Single Table Design

### Phase 1: Enhanced Users Table with Sort Key

#### New Table Structure
```
Table: glad-entities
- Partition Key: PK (String)
- Sort Key: SK (String)
- GSI1-PK: GSI1PK (String)
- GSI1-SK: GSI1SK (String)
```

#### Entity Type: User
```
Item Structure:
{
  PK: "USER#<username>",
  SK: "PROFILE",
  EntityType: "User",
  Username: "<username>",
  Name: "<name>",
  PasswordHash: "<hash>",
  CreatedAt: "<timestamp>",
  UpdatedAt: "<timestamp>"
}
```

**Access Pattern Mapping:**
- AP1 (Get user): GetItem(PK="USER#john", SK="PROFILE")
- AP2 (Create user): PutItem with ConditionExpression on PK+SK
- AP3 (Update user): UpdateItem(PK="USER#john", SK="PROFILE")
- AP4 (User exists): GetItem with ProjectionExpression

### Phase 2: Add User Skills Entity

#### Entity Type: UserSkill
```
Item Structure:
{
  PK: "USER#<username>",
  SK: "SKILL#<skill_name>",
  EntityType: "UserSkill",
  Username: "<username>",
  SkillName: "<skill_name>",
  ProficiencyLevel: "Beginner|Intermediate|Advanced|Expert",
  YearsOfExperience: <number>,
  Endorsements: <number>,
  LastUsedDate: "<date>",
  CreatedAt: "<timestamp>",
  UpdatedAt: "<timestamp>",

  // GSI attributes for querying skills globally
  GSI1PK: "SKILL#<skill_name>",
  GSI1SK: "LEVEL#<proficiency>#USER#<username>"
}
```

#### New Access Patterns with Skills
| Pattern # | Description                 | Type  | Implementation                                                        | Notes                        |
|-----------|-----------------------------|-------|-----------------------------------------------------------------------|------------------------------|
| AP6       | Get all skills for a user   | Read  | Query(PK="USER#john", SK begins_with "SKILL#")                        | Item collection              |
| AP7       | Get specific skill for user | Read  | GetItem(PK="USER#john", SK="SKILL#golang")                            | Direct lookup                |
| AP8       | Add skill to user           | Write | PutItem(PK="USER#john", SK="SKILL#golang")                            | Create skill                 |
| AP9       | Update skill proficiency    | Write | UpdateItem(PK, SK, attributes)                                        | Update skill                 |
| AP10      | Remove skill from user      | Write | DeleteItem(PK="USER#john", SK="SKILL#golang")                         | Delete skill                 |
| AP11      | Find users by skill name    | Read  | Query GSI1(GSI1PK="SKILL#golang")                                     | Cross-user query             |
| AP12      | Find expert users for skill | Read  | Query GSI1(GSI1PK="SKILL#golang", GSI1SK begins_with "LEVEL#Expert#") | Filtered query               |
| AP13      | Get user with all skills    | Read  | Query(PK="USER#john")                                                 | Returns profile + all skills |

### Phase 3: Using Multi-Key Composite GSIs (New DynamoDB Feature)

For more advanced querying, we can leverage the new multi-key composite GSI support:

#### GSI2: Multi-dimensional Skill Queries
```
GSI2 Structure (using composite keys):
- Partition Keys: [SkillName, ProficiencyLevel]
- Sort Keys: [YearsOfExperience, LastUsedDate]

This enables queries like:
"Find users with 'Python' skills at 'Expert' level with 5+ years of experience who used it recently"
```

**Query Pattern:**
```
Query GSI2 where:
  PK1 = "Python" AND
  PK2 = "Expert" AND
  SK1 >= 5 AND
  SK2 >= "2024-01-01"
```

## Design Patterns Applied

### 1. Item Collection Pattern
- **Usage**: User + UserSkills share the same PK
- **Benefit**: Single query retrieves user profile and all skills
- **Example**: Query(PK="USER#john") returns profile + all skills

### 2. Composite Sort Key Pattern (Traditional)
- **Usage**: SK = "SKILL#<name>" enables hierarchical queries
- **Benefit**: Efficient filtering using begins_with, between operators
- **Example**: SK begins_with "SKILL#" returns all skills

### 3. Sparse GSI Pattern
- **Usage**: Only items with skills populate GSI1
- **Benefit**: Reduced GSI storage and write costs
- **Example**: User profiles without skills aren't in skill-lookup GSI

### 4. Natural Key Pattern
- **Usage**: Descriptive keys (USER#, SKILL#) vs generic (PK, SK)
- **Benefit**: Self-documenting, easier debugging
- **Trade-off**: Slightly longer key storage

### 5. Multi-Key Composite GSI Pattern (New Feature)
- **Usage**: Complex multi-dimensional queries
- **Benefit**: No string concatenation, native data types
- **Limitation**: Equality on all PKs, range only on last SK

## Migration Strategy

### Step 1: Table Structure Update (CDK)
1. Rename table from "users" to "glad-entities"
2. Add sort key (SK) as String
3. Add GSI1 with GSI1PK and GSI1SK
4. Add GSI2 with composite multi-keys (optional, future)

### Step 2: Data Migration
1. Scan existing users table
2. Transform each user:
   - PK: "USER#" + username
   - SK: "PROFILE"
   - Keep all existing attributes
3. BatchWrite to new table
4. Verify migration
5. Update application to use new table

### Step 3: Code Refactoring
1. Update models (User, add UserSkill)
2. Update repository layer
3. Update service layer
4. Add new endpoints for skills management
5. Update tests

### Step 4: Skills Feature Implementation
1. Add UserSkill model
2. Add skill repository methods
3. Add skill service methods
4. Add skill API endpoints
5. Add validation and tests

## Protocol for Adding New Entity Types

### Step 1: Define Entity Structure
```go
// Example: Adding "Project" entity

type Project struct {
    ProjectID    string    `json:"project_id" dynamodbav:"project_id"`
    OwnerUsername string   `json:"owner_username" dynamodbav:"owner_username"`
    Name         string    `json:"name" dynamodbav:"name"`
    Description  string    `json:"description" dynamodbav:"description"`
    Status       string    `json:"status" dynamodbav:"status"`
    CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" dynamodbav:"updated_at"`

    // DynamoDB keys
    PK          string `json:"-" dynamodbav:"PK"`
    SK          string `json:"-" dynamodbav:"SK"`
    EntityType  string `json:"entity_type" dynamodbav:"EntityType"`

    // GSI keys (if needed for cross-user queries)
    GSI1PK      string `json:"-" dynamodbav:"GSI1PK,omitempty"`
    GSI1SK      string `json:"-" dynamodbav:"GSI1SK,omitempty"`
}
```

### Step 2: Define Key Pattern
```
Key Pattern Decision Tree:

1. Is this entity owned by a user?
   YES → PK: "USER#<username>"
   NO → PK: "<ENTITY>#<id>"

2. Can a user have multiple of these?
   YES → SK: "<ENTITY>#<unique_id>"
   NO → SK: "PROFILE" or "<ENTITY>"

3. Need to query across all users?
   YES → Add GSI1PK: "<ENTITY>#<attribute>"
   NO → Skip GSI

4. Need multi-dimensional filtering?
   YES → Consider composite GSI2 with up to 4 PKs + 4 SKs
   NO → Use traditional single-key GSI
```

### Step 3: Document Access Patterns
```markdown
| Pattern # | Description | Implementation | RPS |
|-----------|-------------|----------------|-----|
| APxx | Get entity by ID | GetItem(PK, SK) | xxx |
| APxx | List entities for user | Query(PK begins_with) | xxx |
| APxx | Query across users | Query GSI1 | xxx |
```

### Step 4: Add to Table Taxonomy
```markdown
## Entity Types in glad-entities Table

### USER
- PK: USER#<username>
- SK: PROFILE
- Purpose: Core user profile data

### SKILL
- PK: USER#<username>
- SK: SKILL#<skill_name>
- Purpose: User skills and proficiency levels
- GSI1PK: SKILL#<skill_name>
- GSI1SK: LEVEL#<level>#USER#<username>

### PROJECT (example for future)
- PK: USER#<username>
- SK: PROJECT#<project_id>
- Purpose: User-owned projects
- GSI1PK: PROJECT#<project_id>
- GSI1SK: STATUS#<status>#DATE#<created_at>
```

### Step 5: Capacity Planning
```
Formula for new entity:
1. Estimate item size (KB)
2. Estimate write RPS (peak/average)
3. Estimate read RPS (peak/average)
4. Calculate WCU = writes/sec × (item_size_KB / 1)
5. Calculate RCU = reads/sec × (item_size_KB / 4)
6. Check partition limits: 1000 WCU, 3000 RCU per partition
7. If exceeded, consider write sharding
```

## Implementation Checklist

### Phase 1: Foundation (Users with SK)
- [ ] Update CDK stack with new table structure
- [ ] Create migration script
- [ ] Update User model with PK/SK
- [ ] Refactor repository layer
- [ ] Update service layer
- [ ] Update API handlers
- [ ] Add integration tests
- [ ] Deploy and verify
- [ ] Migrate data
- [ ] Monitor and validate

### Phase 2: Add Skills Entity
- [ ] Define UserSkill model
- [ ] Add skill repository methods
- [ ] Add skill service methods
- [ ] Create skill API endpoints
  - POST /users/{username}/skills
  - GET /users/{username}/skills
  - GET /users/{username}/skills/{skillName}
  - PUT /users/{username}/skills/{skillName}
  - DELETE /users/{username}/skills/{skillName}
  - GET /skills/{skillName}/users (GSI query)
- [ ] Add validation logic
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Add API documentation
- [ ] Deploy and verify

### Phase 3: Documentation
- [ ] Create single-table-design.md guide
- [ ] Create entity-addition-protocol.md
- [ ] Create access-patterns-catalog.md
- [ ] Add architecture diagrams
- [ ] Create runbook for common operations

## Cost Analysis

### Current (Simple Table)
```
Assumptions:
- 10,000 users
- 1000 reads/sec (mostly user lookups)
- 100 writes/sec (registrations + updates)
- Average item size: 1 KB

Monthly Costs:
- Reads: 1000 RPS × 2.59M seconds × $0.125/M / 2 (eventually consistent) = $162
- Writes: 100 RPS × 2.59M seconds × $0.625/M = $162
- Storage: 10K users × 1 KB × $0.25/GB = $0.002
Total: ~$324/month
```

### Proposed (Single Table with Skills)
```
Assumptions:
- 10,000 users
- Average 5 skills per user = 50,000 skill items
- 1200 reads/sec (users + skills queries)
- 150 writes/sec (users + skills updates)
- User item: 1 KB, Skill item: 0.5 KB
- GSI1 projection: INCLUDE (user_id, skill_name, level) = 0.3 KB

Base Table:
- Reads: 1200 RPS × 2.59M × $0.125/M / 2 = $194
- Writes: 150 RPS × 2.59M × $0.625/M = $243
- Storage: (10K × 1KB + 50K × 0.5KB) × $0.25/GB = $0.009

GSI1:
- Reads: 100 RPS × 2.59M × $0.125/M / 2 = $16
- Writes: 50 RPS × 2.59M × $0.625/M = $81 (only skill writes)
- Storage: 50K × 0.3KB × $0.25/GB = $0.004

Total: ~$534/month (~65% increase for 5x more entities)
```

## Performance Considerations

### Hot Partition Analysis
```
Current Risk: LOW
- User lookups distributed across 10K usernames
- 1000 RPS / 10,000 users = 0.1 RPS per partition
- Well below 3000 RCU limit

Future Risk with Skills: LOW-MEDIUM
- Skills query via GSI1 could concentrate on popular skills
- Example: "JavaScript" skill queried 100 RPS
- Mitigation: Add random shard suffix for very popular skills
  - GSI1PK: "SKILL#javascript#shard_0" through "SKILL#javascript#shard_9"
```

### Query Optimization Strategies

1. **Batch Operations**: Use BatchGetItem for retrieving multiple users
2. **Projection Expressions**: Only fetch needed attributes
3. **Eventually Consistent Reads**: Use for non-critical paths (2x cheaper)
4. **Pagination**: Implement cursor-based pagination for list operations
5. **Caching**: Add ElastiCache/DAX for frequently accessed users

## Security Considerations

1. **Fine-grained Access Control**: Use IAM policies with condition on PK prefix
   ```json
   {
     "Condition": {
       "ForAllValues:StringLike": {
         "dynamodb:LeadingKeys": ["USER#${cognito:username}"]
       }
     }
   }
   ```

2. **Encryption**: Enable encryption at rest (default in DynamoDB)
3. **VPC Endpoints**: Access DynamoDB via VPC endpoints
4. **Point-in-time Recovery**: Enable PITR for disaster recovery
5. **Backup Strategy**: Automated daily backups with 35-day retention

## Monitoring and Alerting

### Key Metrics to Track
1. **ConsumedReadCapacityUnits** / **ConsumedWriteCapacityUnits**
2. **UserErrors** (throttling, validation errors)
3. **SystemErrors** (service errors)
4. **GetItem/Query/Scan latency** (p50, p99)
5. **GSI throttling events**

### Recommended Alarms
```
1. ThrottledRequests > 10 in 5 minutes
2. UserErrors > 100 in 5 minutes
3. P99 latency > 100ms
4. ConsumedRCU > 80% of provisioned (if not on-demand)
```

## Future Enhancements

### Phase 4: Additional Entity Types (Examples)
1. **Projects**: User projects with skills mapping
2. **Endorsements**: Skill endorsements between users
3. **Certifications**: Professional certifications
4. **Experience**: Work experience entries
5. **Education**: Educational background

### Phase 5: Advanced Features
1. **DynamoDB Streams**: Event-driven architecture
2. **TTL**: Auto-expire temporary data (e.g., session tokens)
3. **Global Tables**: Multi-region replication
4. **Transactions**: Atomic multi-item operations
5. **PartiQL**: SQL-compatible query language

## References

1. [AWS Blog: Multi-key support for GSI](https://aws.amazon.com/blogs/database/multi-key-support-for-global-secondary-index-in-amazon-dynamodb/)
2. [AWS Blog: How Zepto scales with DynamoDB](https://aws.amazon.com/blogs/database/how-zepto-scales-to-millions-of-orders-per-day-using-amazon-dynamodb/)
3. [AWS Blog: Evolve your DynamoDB data model](https://aws.amazon.com/blogs/database/evolve-your-amazon-dynamodb-tables-data-model/)
4. [AWS Best Practices for DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)

## Appendix: Key Patterns Quick Reference

### Pattern 1: Item Collection
```
PK: USER#john
SK: PROFILE          → User entity
SK: SKILL#golang     → Skill entity
SK: SKILL#python     → Skill entity
SK: PROJECT#123      → Project entity
```

### Pattern 2: One-to-Many with GSI
```
Base: PK=ORDER#123, SK=ORDER, customer_id=456
GSI1: PK=CUSTOMER#456, SK=ORDER#123
Query: All orders for customer → Query GSI1(PK=CUSTOMER#456)
```

### Pattern 3: Many-to-Many
```
UserSkillsTable:
  PK: USER#john, SK: SKILL#golang
SkillUsersGSI:
  GSI1PK: SKILL#golang, GSI1SK: USER#john
```

### Pattern 4: Composite Sort Key (Traditional)
```
SK: STATUS#ACTIVE#DATE#2024-01-01#ID#123
Query: begins_with "STATUS#ACTIVE#DATE#2024-01"
```

### Pattern 5: Multi-Key Composite GSI (New Feature)
```
GSI2:
  PK1=SkillName, PK2=Level
  SK1=YearsExp, SK2=LastUsed
Query: Skill="Python" AND Level="Expert" AND YearsExp>=5
```

---

**Document Version**: 1.0
**Last Updated**: 2025-12-07
**Status**: Planning Phase
**Next Review**: After Phase 1 implementation