# Protocol for Adding New Entity Types to Single Table Design

## Quick Reference Guide

This document provides a step-by-step protocol for adding new entity types to the `glad-entities` DynamoDB table using single table design patterns.

## Prerequisites

- Understand the existing table structure (PK, SK, GSI1PK, GSI1SK, GSI2*)
- Have documented access patterns for the new entity
- Know the relationship between new entity and existing entities

## 5-Step Protocol

### Step 1: Analyze Access Patterns

Document all access patterns for your new entity using this template:

```markdown
## Access Patterns for <EntityName>

| Pattern # | Description | Type | Expected RPS | Data Needed |
|-----------|-------------|------|--------------|-------------|
| APxx | [Clear description] | Read/Write | [Peak/Avg] | [Attributes] |

Examples:
- Get entity by ID
- List entities for user
- Query entities by status
- Update entity attributes
- Delete entity
- Cross-user queries (if applicable)
```

### Step 2: Choose Key Pattern

Use this decision tree to determine your PK/SK structure:

```
┌─────────────────────────────────────────┐
│ Is entity owned by a user?              │
└─────────────┬───────────────────────────┘
              │
    ┌─────────┴─────────┐
    │ YES               │ NO
    ▼                   ▼
PK: USER#<username>   PK: <ENTITY>#<id>
    │                   │
    │                   │
┌───┴──────────────────────────────────┐
│ Can user have multiple instances?    │
└───┬──────────────────────────────────┘
    │
    ├── YES: SK: <ENTITY>#<unique_id>
    │        Example: SKILL#golang, PROJECT#123
    │
    └── NO:  SK: <ENTITY> or PROFILE
             Example: SETTINGS, PROFILE

┌─────────────────────────────────────────┐
│ Need to query across all users?         │
└─────────────┬───────────────────────────┘
              │
    ┌─────────┴─────────┐
    │ YES               │ NO
    ▼                   ▼
Add GSI1:            Skip GSI1
GSI1PK: <ENTITY>#<attribute>
GSI1SK: <QUALIFIER>#<value>#USER#<username>

┌─────────────────────────────────────────┐
│ Need multi-dimensional filtering?       │
│ (e.g., by status AND date AND amount)  │
└─────────────┬───────────────────────────┘
              │
    ┌─────────┴─────────┐
    │ YES               │ NO
    ▼                   ▼
Add GSI2 with         Use traditional
composite keys        single-key GSI
(up to 4 PKs + 4 SKs)
```

### Step 3: Define Go Model

Create your entity struct following this template:

```go
package models

import "time"

// <EntityName> represents [description]
type <EntityName> struct {
    // Business attributes
    <Attribute1> string    `json:"<attr1>" dynamodbav:"<attr1>"`
    <Attribute2> string    `json:"<attr2>" dynamodbav:"<attr2>"`
    // ... add all business attributes

    CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" dynamodbav:"updated_at"`

    // DynamoDB system attributes
    PK         string `json:"-" dynamodbav:"PK"`
    SK         string `json:"-" dynamodbav:"SK"`
    EntityType string `json:"entity_type" dynamodbav:"EntityType"`

    // GSI attributes (if needed)
    GSI1PK string `json:"-" dynamodbav:"GSI1PK,omitempty"`
    GSI1SK string `json:"-" dynamodbav:"GSI1SK,omitempty"`

    // GSI2 composite key attributes (if using new multi-key feature)
    GSI2PK1 string `json:"-" dynamodbav:"GSI2PK1,omitempty"`
    GSI2PK2 string `json:"-" dynamodbav:"GSI2PK2,omitempty"`
    GSI2SK1 string `json:"-" dynamodbav:"GSI2SK1,omitempty"`
    GSI2SK2 string `json:"-" dynamodbav:"GSI2SK2,omitempty"`
}

// New<EntityName> creates a new instance with proper key structure
func New<EntityName>(/* params */) (*<EntityName>, error) {
    // Validation
    if /* validation */ {
        return nil, errors.New("validation error")
    }

    now := time.Now()
    entity := &<EntityName>{
        // Set business attributes
        CreatedAt:  now,
        UpdatedAt:  now,
        EntityType: "<EntityName>",
    }

    // Set keys according to your pattern
    entity.SetKeys()

    return entity, nil
}

// SetKeys configures PK, SK, and GSI keys
func (e *<EntityName>) SetKeys() {
    // Base table keys
    e.PK = fmt.Sprintf("USER#%s", e.Username)  // or "<ENTITY>#<id>"
    e.SK = fmt.Sprintf("<ENTITY>#%s", e.ID)

    // GSI1 keys (if needed)
    e.GSI1PK = fmt.Sprintf("<ENTITY>#%s", e.SomeAttribute)
    e.GSI1SK = fmt.Sprintf("TYPE#%s#USER#%s", e.Type, e.Username)

    // GSI2 composite keys (if using multi-key feature)
    // e.GSI2PK1 = e.Attribute1
    // e.GSI2PK2 = e.Attribute2
    // e.GSI2SK1 = fmt.Sprintf("%d", e.NumericValue)
    // e.GSI2SK2 = e.DateValue
}
```

### Step 4: Implement Repository Methods

Add repository methods following this template:

```go
// Create<EntityName> adds a new entity to the table
func (r *DynamoDBRepository) Create<EntityName>(entity *models.<EntityName>) error {
    log := logger.WithComponent("database").With("operation", "Create<EntityName>")
    start := time.Now()

    // Ensure keys are set
    entity.SetKeys()

    item, err := dynamodbattribute.MarshalMap(entity)
    if err != nil {
        log.Error("Failed to marshal entity", "error", err)
        return err
    }

    input := &dynamodb.PutItemInput{
        TableName:           aws.String(TableName),
        Item:                item,
        ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
    }

    _, err = r.client.PutItem(input)
    if err != nil {
        log.Error("Failed to create entity", "error", err)
        return err
    }

    log.Info("Entity created successfully", "duration", time.Since(start))
    return nil
}

// Get<EntityName> retrieves an entity by its keys
func (r *DynamoDBRepository) Get<EntityName>(pk, sk string) (*models.<EntityName>, error) {
    log := logger.WithComponent("database").With("operation", "Get<EntityName>")

    input := &dynamodb.GetItemInput{
        TableName: aws.String(TableName),
        Key: map[string]*dynamodb.AttributeValue{
            "PK": {S: aws.String(pk)},
            "SK": {S: aws.String(sk)},
        },
    }

    result, err := r.client.GetItem(input)
    if err != nil {
        log.Error("Failed to get entity", "error", err)
        return nil, err
    }

    if result.Item == nil {
        return nil, errors.New("entity not found")
    }

    var entity models.<EntityName>
    err = dynamodbattribute.UnmarshalMap(result.Item, &entity)
    if err != nil {
        log.Error("Failed to unmarshal entity", "error", err)
        return nil, err
    }

    return &entity, nil
}

// List<EntityName>ForUser retrieves all entities for a user (item collection query)
func (r *DynamoDBRepository) List<EntityName>ForUser(username string) ([]*models.<EntityName>, error) {
    log := logger.WithComponent("database").With("operation", "List<EntityName>ForUser")

    input := &dynamodb.QueryInput{
        TableName:              aws.String(TableName),
        KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":pk":        {S: aws.String(fmt.Sprintf("USER#%s", username))},
            ":sk_prefix": {S: aws.String("<ENTITY>#")},
        },
    }

    result, err := r.client.Query(input)
    if err != nil {
        log.Error("Failed to query entities", "error", err)
        return nil, err
    }

    var entities []*models.<EntityName>
    for _, item := range result.Items {
        var entity models.<EntityName>
        if err := dynamodbattribute.UnmarshalMap(item, &entity); err != nil {
            log.Error("Failed to unmarshal entity", "error", err)
            continue
        }
        entities = append(entities, &entity)
    }

    return entities, nil
}

// Query<EntityName>ByGSI queries entities using GSI1
func (r *DynamoDBRepository) Query<EntityName>ByGSI(gsi1pk string, gsi1skPrefix string) ([]*models.<EntityName>, error) {
    log := logger.WithComponent("database").With("operation", "Query<EntityName>ByGSI")

    input := &dynamodb.QueryInput{
        TableName:              aws.String(TableName),
        IndexName:              aws.String("GSI1"),
        KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :sk)"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":pk": {S: aws.String(gsi1pk)},
            ":sk": {S: aws.String(gsi1skPrefix)},
        },
    }

    result, err := r.client.Query(input)
    if err != nil {
        log.Error("Failed to query GSI", "error", err)
        return nil, err
    }

    var entities []*models.<EntityName>
    for _, item := range result.Items {
        var entity models.<EntityName>
        if err := dynamodbattribute.UnmarshalMap(item, &entity); err != nil {
            log.Error("Failed to unmarshal entity", "error", err)
            continue
        }
        entities = append(entities, &entity)
    }

    return entities, nil
}

// Update<EntityName> updates an existing entity
func (r *DynamoDBRepository) Update<EntityName>(entity *models.<EntityName>) error {
    log := logger.WithComponent("database").With("operation", "Update<EntityName>")

    entity.UpdatedAt = time.Now()
    entity.SetKeys() // Ensure keys are current

    item, err := dynamodbattribute.MarshalMap(entity)
    if err != nil {
        log.Error("Failed to marshal entity", "error", err)
        return err
    }

    input := &dynamodb.PutItemInput{
        TableName:           aws.String(TableName),
        Item:                item,
        ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
    }

    _, err = r.client.PutItem(input)
    if err != nil {
        log.Error("Failed to update entity", "error", err)
        return err
    }

    return nil
}

// Delete<EntityName> removes an entity from the table
func (r *DynamoDBRepository) Delete<EntityName>(pk, sk string) error {
    log := logger.WithComponent("database").With("operation", "Delete<EntityName>")

    input := &dynamodb.DeleteItemInput{
        TableName: aws.String(TableName),
        Key: map[string]*dynamodb.AttributeValue{
            "PK": {S: aws.String(pk)},
            "SK": {S: aws.String(sk)},
        },
        ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
    }

    _, err := r.client.DeleteItem(input)
    if err != nil {
        log.Error("Failed to delete entity", "error", err)
        return err
    }

    return nil
}
```

### Step 5: Update Table Taxonomy Documentation

Add your entity to the table taxonomy document:

```markdown
## <EntityName> Entity

### Purpose
[Brief description of what this entity represents and why it exists]

### Key Structure
```
PK: <pattern>
SK: <pattern>
EntityType: "<EntityName>"

# GSI1 (if applicable)
GSI1PK: <pattern>
GSI1SK: <pattern>

# GSI2 (if using composite multi-key)
GSI2PK1: <attribute>
GSI2PK2: <attribute>
GSI2SK1: <attribute>
GSI2SK2: <attribute>
```

### Example Items
```json
{
  "PK": "USER#john",
  "SK": "<ENTITY>#<id>",
  "EntityType": "<EntityName>",
  // ... business attributes
  "CreatedAt": "2024-01-01T00:00:00Z",
  "UpdatedAt": "2024-01-01T00:00:00Z"
}
```

### Access Patterns

| Pattern # | Description | Implementation | Notes |
|-----------|-------------|----------------|-------|
| APxx | [Description] | [DynamoDB operation] | [Special considerations] |

### Relationships
- **Parent**: [If applicable]
- **Children**: [If applicable]
- **Related Entities**: [Cross-references]

### Capacity Estimates
- **Average Item Size**: X KB
- **Expected Items**: X
- **Read RPS**: X (peak), X (avg)
- **Write RPS**: X (peak), X (avg)
- **Storage**: ~X KB total

### Special Considerations
- [Any unique aspects, limitations, or important notes]
```

## Checklist for Adding New Entity

Use this checklist to ensure you've completed all steps:

### Planning Phase
- [ ] Document all access patterns
- [ ] Determine key structure (PK, SK, GSI keys)
- [ ] Calculate capacity requirements
- [ ] Check for hot partition risks
- [ ] Review with team/lead

### Implementation Phase
- [ ] Create Go model with proper struct tags
- [ ] Implement `New<EntityName>()` constructor
- [ ] Implement `SetKeys()` method
- [ ] Add validation logic
- [ ] Create repository methods:
  - [ ] Create
  - [ ] Get
  - [ ] List/Query
  - [ ] Update
  - [ ] Delete
- [ ] Add service layer methods
- [ ] Create API handlers/endpoints
- [ ] Add input validation
- [ ] Implement error handling

### Testing Phase
- [ ] Write unit tests for model
- [ ] Write unit tests for repository
- [ ] Write unit tests for service
- [ ] Write integration tests
- [ ] Test GSI queries (if applicable)
- [ ] Test error cases
- [ ] Performance test with realistic data

### Documentation Phase
- [ ] Update table taxonomy
- [ ] Document access patterns
- [ ] Add code examples
- [ ] Update API documentation
- [ ] Add runbook entries

### Deployment Phase
- [ ] Update CDK if GSI changes needed
- [ ] Deploy infrastructure changes
- [ ] Deploy application code
- [ ] Monitor metrics
- [ ] Verify functionality in production

## Common Patterns Reference

### Pattern 1: User-Owned Entity (One-to-Many)
```
Example: User Skills
PK: USER#john
SK: SKILL#golang
GSI1PK: SKILL#golang
GSI1SK: LEVEL#Expert#USER#john
```

### Pattern 2: Global Entity (Not User-Owned)
```
Example: Products
PK: PRODUCT#123
SK: PRODUCT
GSI1PK: CATEGORY#electronics
GSI1SK: PRICE#999
```

### Pattern 3: Relationship Entity (Many-to-Many)
```
Example: User-Project Membership
PK: USER#john
SK: PROJECT#456
GSI1PK: PROJECT#456
GSI1SK: USER#john
```

### Pattern 4: Hierarchical Data
```
Example: Course Lessons
PK: COURSE#101
SK: LESSON#1
SK: LESSON#2
SK: LESSON#3
```

### Pattern 5: Sparse GSI (Filtered Queries)
```
Example: Active Subscriptions Only
PK: USER#john
SK: SUBSCRIPTION
Status: ACTIVE  ← Only items with Status populate GSI
GSI1PK: STATUS#ACTIVE
GSI1SK: EXPIRES#2024-12-31
```

### Pattern 6: Composite Multi-Key GSI (New Feature)
```
Example: Advanced Product Search
Base Table:
  PK: PRODUCT#123
  SK: PRODUCT

GSI2 (Composite Keys):
  PK1: Category (string)
  PK2: Brand (string)
  SK1: Price (number)
  SK2: Rating (number)

Query:
  Category="Electronics" AND
  Brand="Apple" AND
  Price >= 500 AND
  Rating >= 4.5
```

## Capacity Planning Formulas

```
Item Size Calculation:
- Base overhead: ~100 bytes per item
- Attribute overhead: ~1 byte per attribute name
- String: length in bytes (UTF-8)
- Number: ~variable (1-21 bytes)
- Boolean: 1 byte
- List/Map: sum of elements + overhead

Total Item Size = 100 + Σ(attribute_name_length + attribute_value_size)

RCU Calculation (Eventually Consistent):
RCU = (reads_per_sec × item_size_KB) / 8
RCU = (reads_per_sec × item_size_KB) / 4 (Strongly Consistent)

WCU Calculation:
WCU = (writes_per_sec × item_size_KB) / 1

Partition Throughput Limits:
- Max 3,000 RCU per partition
- Max 1,000 WCU per partition

Hot Partition Check:
If (entity_write_RPS / distinct_partition_keys) > 1000:
  → Implement write sharding

If (entity_read_RPS / distinct_partition_keys) > 3000:
  → Add caching layer or distribute reads
```

## Troubleshooting Guide

### Issue: Hot Partition
**Symptoms**: Throttling errors, high latency
**Solutions**:
1. Add write sharding: `PK: <base>#<shard_id>`
2. Use composite keys to distribute load
3. Add caching layer (DAX, ElastiCache)

### Issue: Expensive GSI
**Symptoms**: High write costs, storage costs
**Solutions**:
1. Use sparse GSI (only index subset)
2. Change projection to KEYS_ONLY or INCLUDE
3. Consider denormalization instead

### Issue: Complex Queries
**Symptoms**: Multiple DynamoDB calls, slow performance
**Solutions**:
1. Use item collections (same PK)
2. Add GSI for cross-user queries
3. Consider composite multi-key GSI
4. Denormalize frequently accessed data

### Issue: Large Item Size
**Symptoms**: Item > 400KB limit
**Solutions**:
1. Split into multiple items (item collection)
2. Store large attributes in S3, reference in DynamoDB
3. Use compression for large text fields

## Best Practices Summary

1. **Keys should be descriptive**: Use `USER#john` not `PK=abc123`
2. **Group related entities**: Use item collections when access correlation > 50%
3. **Project minimally**: Only include attributes you query in GSI
4. **Estimate capacity**: Calculate RCU/WCU before implementation
5. **Monitor from day one**: Set up alarms for throttling
6. **Document everything**: Future you will thank present you
7. **Test with realistic data**: Production data patterns matter
8. **Consider write amplification**: Each GSI doubles write cost
9. **Use sparse GSIs**: When querying <50% of items
10. **Leverage new composite keys**: For multi-dimensional queries

## Getting Help

- Review the main planning document: `docs/dynamodb-single-table-design-plan.md`
- Check AWS documentation: https://docs.aws.amazon.com/dynamodb/
- Consult with team lead before adding GSIs
- Test queries in DynamoDB Local first

---

**Document Version**: 1.0
**Last Updated**: 2025-12-07
**Maintained By**: GLAD Engineering Team