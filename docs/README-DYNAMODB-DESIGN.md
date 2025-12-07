# DynamoDB Single Table Design - Complete Solution

## Overview

This directory contains a comprehensive solution for implementing DynamoDB single table design in the GLAD project, including support for the new **multi-key composite GSI feature** (up to 4 partition keys + 4 sort keys).

## 📋 Documentation Files

### 1. **Main Planning Document**
📄 [`dynamodb-single-table-design-plan.md`](./dynamodb-single-table-design-plan.md)

**Purpose**: Comprehensive implementation plan covering all phases

**Contents**:
- Current state analysis
- Proposed single table design (3 phases)
- Entity designs (User, UserSkill with examples)
- Access pattern mapping (13+ patterns)
- Migration strategy
- Cost analysis and capacity planning
- Security, monitoring, and best practices

**Use this for**: Understanding the big picture, planning phases, and architectural decisions

---

### 2. **Entity Addition Protocol**
📄 [`entity-addition-protocol.md`](./entity-addition-protocol.md)

**Purpose**: Step-by-step guide for adding new entity types

**Contents**:
- 5-step protocol with decision trees
- Visual flowcharts for key pattern selection
- Go code templates for models and repository
- Complete checklist for implementation
- Common patterns reference
- Capacity planning formulas
- Troubleshooting guide

**Use this for**: Adding new entities (Projects, Settings, etc.) to the single table

---

### 3. **Quick Reference Guide**
📄 [`dynamodb-quick-reference.md`](./dynamodb-quick-reference.md)

**Purpose**: Cheat sheet for daily development

**Contents**:
- Table structure overview
- Key patterns cheat sheet
- Common query examples
- Composite multi-key GSI rules (8-key support)
- Cost comparison (traditional vs composite)
- Capacity planning formulas
- Anti-patterns to avoid
- AWS CLI commands
- Monitoring metrics

**Use this for**: Daily development, quick lookups, and reference during coding

---

### 4. **Skill File with Examples**
📄 [`../.claude/skills/dynamo-db-single-table-design.md`](../.claude/skills/dynamo-db-single-table-design.md)

**Purpose**: Concrete examples demonstrating all patterns

**Contents**:
- Example 1: User Profile + Skills (Item Collection)
- Example 2: Multi-Key Composite GSI (8-key feature)
- Example 3: Adding Projects entity
- Example 4: Capacity planning & cost estimation
- Example 5: Migration from multi-table to single table

**Use this for**: Learning by example, understanding patterns in practice

---

## 🚀 Quick Start

### Phase 1: Understand Current State
1. Read the "Current State Analysis" section in the main planning document
2. Review existing code:
   - `/cmd/app/internal/database/dynamodb.go` - Current repository
   - `/cmd/app/internal/models/user.go` - User model
   - `/deployments/app/cdk.go` - Infrastructure

### Phase 2: Design Your First Entity (Skills)
1. Review Example 1 in the skill file
2. Follow the Entity Addition Protocol for UserSkill
3. Use the Quick Reference for query patterns

### Phase 3: Implement
1. Update CDK stack (add SK, GSI1, optionally GSI2)
2. Create UserSkill model
3. Add repository methods
4. Add service layer
5. Create API endpoints
6. Test thoroughly

## 🎯 Key Concepts

### Single Table Design Benefits
- ✅ **Single Query Efficiency**: Get user + all skills in one query
- ✅ **Cost Optimization**: 1 table vs multiple tables
- ✅ **Better Performance**: Data locality, fewer round trips
- ✅ **Simpler Operations**: One table to manage, backup, monitor

### Multi-Key Composite GSI (New Feature!)
DynamoDB now supports **up to 4 partition keys + 4 sort keys** in GSIs:

```
GSI2 Example:
  PK1: SkillName (String)
  PK2: ProficiencyLevel (String)
  SK1: YearsOfExperience (Number) ← Native number type!
  SK2: LastUsedDate (String)

Query:
  Find Python experts with 5+ years used recently
  PK1="Python" AND PK2="Expert" AND SK1>=5 AND SK2>="2024-01-01"
```

**Benefits**:
- Native data types (no concatenation!)
- Type-safe queries
- Better maintainability
- More flexible filtering

## 📊 Entity Design Patterns

### Pattern 1: Item Collection (User-Owned Entities)
```
PK: USER#john, SK: PROFILE        → User
PK: USER#john, SK: SKILL#golang   → Skill
PK: USER#john, SK: SKILL#python   → Skill
PK: USER#john, SK: PROJECT#123    → Project

Query(PK="USER#john") → Returns ALL entities for user!
```

### Pattern 2: Cross-User Queries (GSI1)
```
Base Table:
  PK: USER#john, SK: SKILL#golang

GSI1:
  GSI1PK: SKILL#golang, GSI1SK: LEVEL#Expert#USER#john

Query GSI1(GSI1PK="SKILL#golang") → Find all users with skill
```

### Pattern 3: Multi-Dimensional Queries (GSI2 Composite)
```
GSI2 with 4 PKs + 4 SKs for complex filtering
Example: Find active, high-priority projects due before date
```

## 🔧 Implementation Checklist

### Step 1: Infrastructure (CDK)
- [ ] Add sort key (SK) to table
- [ ] Add GSI1 (GSI1PK, GSI1SK)
- [ ] Add GSI2 with composite keys (optional)
- [ ] Update table name to `glad-entities`
- [ ] Deploy infrastructure

### Step 2: Models
- [ ] Update User model with PK/SK fields
- [ ] Create UserSkill model
- [ ] Add SetKeys() methods
- [ ] Add validation

### Step 3: Repository Layer
- [ ] Update user repository methods
- [ ] Create skill repository methods
- [ ] Add GSI query methods
- [ ] Add error handling

### Step 4: Service Layer
- [ ] Update user service
- [ ] Create skill service
- [ ] Add business logic validation

### Step 5: API Layer
- [ ] Create skill endpoints
- [ ] Update API documentation
- [ ] Add request/response DTOs

### Step 6: Testing
- [ ] Unit tests for models
- [ ] Unit tests for repository
- [ ] Integration tests
- [ ] Load testing

### Step 7: Migration
- [ ] Create migration script
- [ ] Test on sample data
- [ ] Execute migration
- [ ] Verify data integrity
- [ ] Monitor performance

## 💰 Cost Estimation

For 10,000 users with 5 skills each:

| Component | Monthly Cost |
|-----------|-------------|
| Base Table (reads) | $324 |
| Base Table (writes) | $324 |
| GSI1 (reads) | $81 |
| GSI1 (writes) | $243 |
| Storage | ~$0.02 |
| **Total** | **~$972/month** |

**Optimization Tips**:
- Use eventually consistent reads (50% savings) ✅
- Add caching for hot users (30% RPS reduction)
- Use sparse GSIs (already applied) ✅
- Batch operations where possible

## 📈 Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| GetItem latency | < 10ms | P99 |
| Query latency | < 20ms | P99 |
| Throughput | 1000+ RPS | Per table |
| Availability | 99.99% | DynamoDB SLA |

## 🔍 Monitoring

### Key Metrics
- ConsumedReadCapacityUnits
- ConsumedWriteCapacityUnits
- UserErrors (throttling)
- P50, P99 latency

### Critical Alarms
```
ThrottledRequests > 0
UserErrors > 100 in 5 min
P99 latency > 100ms
ConsumedRCU > 80% of provisioned
```

## 🚨 Common Pitfalls to Avoid

❌ **Don't**: Use Scan operations for queries
✅ **Do**: Use Query with proper key conditions

❌ **Don't**: Create generic keys (PK, SK, GSI1PK)
✅ **Do**: Use descriptive keys (USER#john, SKILL#golang)

❌ **Don't**: Over-normalize (separate table per entity)
✅ **Do**: Use item collections for related entities

❌ **Don't**: Use mutable attributes as GSI keys
✅ **Do**: Use stable attributes or accept write amplification

❌ **Don't**: Store everything in one item
✅ **Do**: Use item collections with separate items

## 📚 Additional Resources

### AWS Documentation
- [Multi-Key GSI Support](https://aws.amazon.com/blogs/database/multi-key-support-for-global-secondary-index-in-amazon-dynamodb/)
- [Zepto's DynamoDB Architecture](https://aws.amazon.com/blogs/database/how-zepto-scales-to-millions-of-orders-per-day-using-amazon-dynamodb/)
- [Evolving DynamoDB Data Models](https://aws.amazon.com/blogs/database/evolve-your-amazon-dynamodb-tables-data-model/)
- [SQL to NoSQL Migration](https://aws.amazon.com/blogs/database/sql-to-nosql-planning-your-application-migration-to-amazon-dynamodb/)

### Tools
- **DynamoDB Local**: Local testing environment
- **NoSQL Workbench**: Visual data modeling
- **AWS CLI**: Command-line operations
- **AWS SDK for Go**: v2.x with improved DynamoDB support

## 🎓 Learning Path

1. **Beginner**: Read Example 1 in skill file (User + Skills)
2. **Intermediate**: Review Quick Reference, understand access patterns
3. **Advanced**: Study composite multi-key GSI (Example 2)
4. **Expert**: Read full planning document, understand cost optimization

## 🤝 Contributing

When adding new entity types:
1. Follow the Entity Addition Protocol
2. Document access patterns
3. Update this README with new examples
4. Add capacity planning estimates
5. Update monitoring dashboards

## 📞 Support

- **Planning Questions**: Review main planning document
- **Implementation Help**: Use Entity Addition Protocol
- **Quick Lookups**: Use Quick Reference Guide
- **Learning**: Review examples in skill file

## 🏗️ Project Structure

```
docs/
├── README-DYNAMODB-DESIGN.md          (this file)
├── dynamodb-single-table-design-plan.md  (main planning doc)
├── entity-addition-protocol.md        (step-by-step guide)
└── dynamodb-quick-reference.md        (daily reference)

.claude/skills/
└── dynamo-db-single-table-design.md   (examples)

cmd/app/internal/
├── database/
│   └── dynamodb.go                    (repository layer)
├── models/
│   ├── user.go                        (user model)
│   └── user_skill.go                  (skill model - to be added)
└── service/
    ├── user_service.go                (user service)
    └── skill_service.go               (skill service - to be added)

deployments/app/
└── cdk.go                             (infrastructure)
```

## ✅ Next Steps

1. **Review**: Read this README and the main planning document
2. **Understand**: Study Example 1 (User + Skills) in the skill file
3. **Plan**: Review Phase 1 implementation checklist
4. **Discuss**: Align with team on approach and timeline
5. **Implement**: Start with CDK changes, then models, then repository
6. **Test**: Write comprehensive tests before migration
7. **Deploy**: Execute migration during low-traffic window
8. **Monitor**: Watch metrics closely after deployment

---

**Last Updated**: 2025-12-07
**Version**: 1.0
**Maintained By**: GLAD Engineering Team

**Questions?** Review the documentation files linked above or consult the AWS resources.