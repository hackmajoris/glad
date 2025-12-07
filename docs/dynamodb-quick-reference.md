# DynamoDB Single Table Design - Quick Reference

## Table Structure Overview

```
Table: glad-entities
- Partition Key: PK (String)
- Sort Key: SK (String)
- GSI1-PK: GSI1PK (String)
- GSI1-SK: GSI1SK (String)
- GSI2: Up to 4 PKs + 4 SKs (composite multi-key support)
```

## Current Entity Types

### User Profile
```
PK: USER#<username>
SK: PROFILE
EntityType: User
```

### User Skill
```
PK: USER#<username>
SK: SKILL#<skill_name>
EntityType: UserSkill

GSI1PK: SKILL#<skill_name>
GSI1SK: LEVEL#<proficiency>#USER#<username>
```

## Key Patterns Cheat Sheet

| Pattern       | PK            | SK             | GSI1PK          | GSI1SK                   | Use Case        |
|---------------|---------------|----------------|-----------------|--------------------------|-----------------|
| User Profile  | `USER#john`   | `PROFILE`      | -               | -                        | User data       |
| User Skill    | `USER#john`   | `SKILL#golang` | `SKILL#golang`  | `LEVEL#Expert#USER#john` | Skills per user |
| User Project  | `USER#john`   | `PROJECT#123`  | `PROJECT#123`   | `USER#john`              | Projects        |
| Setting       | `USER#john`   | `SETTINGS`     | -               | -                        | User settings   |
| Global Entity | `PRODUCT#123` | `PRODUCT`      | `CATEGORY#tech` | `PRICE#999`              | Products        |

## Common Query Examples

### Get User Profile
```go
GetItem(
  PK: "USER#john",
  SK: "PROFILE"
)
```

### Get User with All Skills
```go
Query(
  PK: "USER#john"
)
// Returns: PROFILE + all SKILL#* items
```

### Get Specific Skill
```go
GetItem(
  PK: "USER#john",
  SK: "SKILL#golang"
)
```

### Find All Users with Skill
```go
Query(
  IndexName: "GSI1",
  GSI1PK: "SKILL#golang"
)
```

### Find Expert Users for Skill
```go
Query(
  IndexName: "GSI1",
  GSI1PK: "SKILL#golang",
  GSI1SK: begins_with("LEVEL#Expert#")
)
```

## Composite Multi-Key GSI (New Feature - Up to 8 Keys!)

DynamoDB now supports up to **4 partition keys** and **4 sort keys** in composite GSIs.

### Rules
1. **Equality (=)** required on ALL partition key attributes
2. **Range operators** (<, >, BETWEEN) only on the LAST sort key
3. **Cannot skip** sort keys (must use left-to-right)
4. **Native data types** - no concatenation needed!

### Example: Advanced Skill Search

```go
// GSI2 Definition
GSI2 Composite Keys:
  PK1: SkillName (String)
  PK2: ProficiencyLevel (String)
  SK1: YearsOfExperience (Number)
  SK2: LastUsedDate (String)

// Query: Find Python experts with 5+ years who used it recently
Query(
  IndexName: "GSI2",
  GSI2PK1: "Python",           // Equality
  GSI2PK2: "Expert",           // Equality
  GSI2SK1: >= 5,               // Range (not last, so must use >=)
  GSI2SK2: >= "2024-01-01"     // Range on last SK
)
```

### When to Use Multi-Key Composite GSI

✅ **Use when**:
- Need to filter by 3+ dimensions
- Want type-safe queries (no string parsing)
- Complex business queries across multiple attributes
- Need efficient multi-dimensional searches

❌ **Don't use when**:
- Simple 1-2 dimension queries (use traditional GSI)
- Attributes change frequently (write amplification)
- Low query volume (not worth complexity)

## Cost Comparison

### Traditional Concatenated SK
```
Item: {
  PK: "SKILL#golang",
  SK: "LEVEL#Expert#EXP#5#DATE#2024-01-01#USER#john"
}

Drawbacks:
- String parsing required
- No type safety for numbers/dates
- Complex to maintain
- Range queries limited
```

### Composite Multi-Key
```
Item: {
  GSI2PK1: "golang",      // String
  GSI2PK2: "Expert",      // String
  GSI2SK1: 5,             // Number (native!)
  GSI2SK2: "2024-01-01",  // String
  GSI2SK3: "john"         // String
}

Benefits:
+ Native data types
+ No string parsing
+ Type-safe queries
+ Better maintainability
```

## Capacity Planning Quick Reference

```
RCU (Eventually Consistent) = (reads/sec × KB) ÷ 8
RCU (Strongly Consistent)   = (reads/sec × KB) ÷ 4
WCU                         = (writes/sec × KB) ÷ 1

Partition Limits:
- 3,000 RCU per partition
- 1,000 WCU per partition
- 10 GB per partition

Hot Partition Threshold:
- Write RPS per distinct PK > 1000 → Add sharding
- Read RPS per distinct PK > 3000 → Add caching
```

## Item Size Guidelines

```
Target: < 10 KB per item (optimal)
Warning: 10-100 KB (acceptable, watch costs)
Critical: > 100 KB (consider S3 offload)
Maximum: 400 KB (hard DynamoDB limit)
```

## GSI Projection Types

| Type      | Storage Cost | Read Cost | Query Latency | Use When                                 |
|-----------|--------------|-----------|---------------|------------------------------------------|
| KEYS_ONLY | Lowest       | Lowest*   | Higher        | Need PK/SK only, can GetItem for details |
| INCLUDE   | Medium       | Medium    | Medium        | Need 2-3 specific attributes             |
| ALL       | Highest      | Highest   | Lowest        | Need most attributes, avoid GetItem      |

\* Requires additional GetItem if you need non-key attributes

## Error Handling Quick Reference

### ProvisionedThroughputExceededException
- **Cause**: Hot partition, insufficient capacity
- **Fix**: Add exponential backoff, enable auto-scaling, add sharding

### ConditionalCheckFailedException
- **Cause**: Condition not met (item exists, version mismatch)
- **Fix**: Expected in normal operations, handle gracefully

### ValidationException
- **Cause**: Invalid request (missing keys, wrong types)
- **Fix**: Validate input before DynamoDB call

### ItemCollectionSizeLimitExceededException
- **Cause**: Item collection > 10 GB
- **Fix**: Redistribute items across more partition keys

## Access Pattern Complexity Matrix

| Complexity   | Pattern                  | Solution                               | Example                    |
|--------------|--------------------------|----------------------------------------|----------------------------|
| Simple       | Single item lookup       | GetItem                                | Get user by username       |
| Medium       | Items for one PK         | Query on base table                    | Get all skills for user    |
| Medium+      | Items across PKs         | Query on GSI                           | Find all users with skill  |
| Complex      | Multi-dimensional filter | Composite multi-key GSI                | Experts with 5+ years      |
| Very Complex | Cross-table aggregation  | Consider separate table or denormalize | User stats across entities |

## Decision Trees

### Should I Add a GSI?

```
Need to query by non-key attribute?
├─ YES: Can you use base table SK?
│   ├─ YES: Use SK patterns (hierarchical, composite)
│   └─ NO: Need to query across different PKs?
│       ├─ YES: Add GSI
│       └─ NO: Filter in application (if < 1MB result)
└─ NO: Use base table
```

### Should I Use Composite Multi-Key GSI?

```
Need to filter by how many attributes?
├─ 1-2 attributes: Use traditional GSI with composite SK
├─ 3-4 attributes: Consider composite multi-key GSI
│   ├─ Attributes change frequently?
│   │   ├─ YES: Stick with traditional (avoid write amplification)
│   │   └─ NO: Use composite multi-key GSI
│   └─ Need range queries on multiple attributes?
│       ├─ YES: Use composite multi-key (range on last only)
│       └─ NO: Traditional GSI is simpler
└─ 5+ attributes: Consider denormalization or separate table
```

### Should I Denormalize?

```
How often is data accessed together?
├─ > 80%: Strong denormalization candidate
│   └─ How often does it change?
│       ├─ Rarely: Denormalize ✅
│       └─ Frequently: Keep normalized, use GSI
├─ 50-80%: Calculate cost
│   └─ (Read_RPS × 2) > (Write_RPS × N_copies)?
│       ├─ YES: Denormalize
│       └─ NO: Keep normalized
└─ < 50%: Keep normalized
```

## Common Anti-Patterns to Avoid

❌ **Scan instead of Query**
```go
// BAD
Scan(TableName: "glad-entities")

// GOOD
Query(PK: "USER#john", SK: begins_with("SKILL#"))
```

❌ **Over-normalized (too many tables)**
```
// BAD: Separate table for each entity
users-table, skills-table, projects-table, settings-table

// GOOD: Single table with entity types
glad-entities (with PK/SK patterns)
```

❌ **Under-normalized (god object)**
```go
// BAD: Everything in one item
{
  PK: "USER#john",
  SK: "PROFILE",
  Skills: [...100 skills...],
  Projects: [...50 projects...],
  Settings: {...}
}

// GOOD: Item collection
PK: USER#john, SK: PROFILE
PK: USER#john, SK: SKILL#golang
PK: USER#john, SK: SKILL#python
```

❌ **Mutable GSI keys**
```go
// BAD: Using frequently changing attribute as GSI key
GSI1PK: "STATUS#" + currentStatus  // Changes every hour!

// GOOD: Use stable attributes or accept write amplification
GSI1PK: "SKILL#" + skillName  // Rarely changes
```

❌ **Generic key names**
```go
// BAD
PK: "123"
SK: "456"
GSI1PK: "abc"

// GOOD
PK: "USER#john"
SK: "SKILL#golang"
GSI1PK: "SKILL#golang"
```

## Performance Optimization Tips

1. **Batch Operations**: Use BatchGetItem/BatchWriteItem for multiple items
2. **Eventually Consistent Reads**: Use for non-critical paths (50% cost reduction)
3. **Projection Expressions**: Only fetch attributes you need
4. **Pagination**: Limit queries, use LastEvaluatedKey
5. **Caching**: Add DAX or ElastiCache for hot data
6. **Parallel Queries**: Query multiple partitions concurrently
7. **Local Testing**: Use DynamoDB Local for development

## Monitoring Metrics

### Critical Alarms
```
ConsumedReadCapacityUnits > 80% of provisioned
ConsumedWriteCapacityUnits > 80% of provisioned
UserErrors > 100 in 5 minutes
SystemErrors > 10 in 5 minutes
ThrottledRequests > 0
```

### Key Metrics to Track
- P50, P99 latency for GetItem, Query, PutItem
- Consumed vs Provisioned capacity
- Item count and table size
- GSI backfill progress (if adding new GSI)

## Tools and Resources

### Development Tools
- **DynamoDB Local**: Local testing environment
- **NoSQL Workbench**: Visual data modeling tool
- **AWS CLI**: Command-line operations
- **AWS SDK**: Go, Python, Java, JavaScript SDKs

### Useful AWS CLI Commands
```bash
# Query table
aws dynamodb query --table-name glad-entities \
  --key-condition-expression "PK = :pk" \
  --expression-attribute-values '{":pk":{"S":"USER#john"}}'

# Get item
aws dynamodb get-item --table-name glad-entities \
  --key '{"PK":{"S":"USER#john"},"SK":{"S":"PROFILE"}}'

# Scan table (use sparingly!)
aws dynamodb scan --table-name glad-entities \
  --filter-expression "EntityType = :type" \
  --expression-attribute-values '{":type":{"S":"UserSkill"}}'

# Describe table
aws dynamodb describe-table --table-name glad-entities

# Update item
aws dynamodb update-item --table-name glad-entities \
  --key '{"PK":{"S":"USER#john"},"SK":{"S":"SKILL#golang"}}' \
  --update-expression "SET ProficiencyLevel = :level" \
  --expression-attribute-values '{":level":{"S":"Expert"}}'
```

## Testing Checklist

- [ ] Test with realistic item sizes
- [ ] Test pagination for large result sets
- [ ] Test conditional operations
- [ ] Test GSI queries
- [ ] Test error handling (throttling, not found, etc.)
- [ ] Load test with expected RPS
- [ ] Test hot partition scenarios
- [ ] Verify monitoring and alarms

## Migration Checklist

- [ ] Backup existing data
- [ ] Create new table structure
- [ ] Test migration script on sample data
- [ ] Run migration during low-traffic window
- [ ] Verify data integrity
- [ ] Update application code
- [ ] Monitor for errors
- [ ] Keep old table as backup (don't delete immediately)

---

**Quick Links**:
- [Full Planning Document](./dynamodb-single-table-design-plan.md)
- [Entity Addition Protocol](./entity-addition-protocol.md)
- [AWS DynamoDB Documentation](https://docs.aws.amazon.com/dynamodb/)
- [AWS Blog: Multi-Key GSI](https://aws.amazon.com/blogs/database/multi-key-support-for-global-secondary-index-in-amazon-dynamodb/)

**Last Updated**: 2025-12-07
