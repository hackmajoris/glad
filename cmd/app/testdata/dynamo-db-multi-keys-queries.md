
# Get all Expert Python users with 5+ years:
  aws dynamodb query \
    --table-name glad-entities-production \
    --index-name SkillsByLevel \
    --key-condition-expression "SkillName = :skill AND ProficiencyLevel = :level AND YearsOfExperience >= :years" \
    --expression-attribute-values '{
      ":skill": {"S": "Python"},
      ":level": {"S": "Expert"},
      ":years": {"N": "5"}
    }'

# Get user profile + all skills:
 aws dynamodb query \
   --table-name glad-entities-production \
   --index-name ByUser \
	--profile passbrains-ilisa-amplify \
	--key-condition-expression "Username = :user" \
    --expression-attribute-values '{":user": {"S": "john"}}'

# Get just user profile:
 aws dynamodb query \
   --table-name glad-entities-production \
   --profile passbrains-ilisa-amplify \
   --index-name ByUser \
   --key-condition-expression "Username = :user AND EntityType = :type" \
   --expression-attribute-values '{
     ":user": {"S": "john"},
     ":type": {"S": "User"}
   }'


# User 1: John Doe
aws dynamodb put-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--item '{
"entity_id": {"S": "USER-john"},
"EntityType": {"S": "User"},
"Username": {"S": "john"},
"Name": {"S": "John Doe"},
"Email": {"S": "john@example.com"},
"PasswordHash": {"S": "$2a$10$examplehash"},
"CreatedAt": {"S": "2025-01-01T10:00:00Z"},
"UpdatedAt": {"S": "2025-01-01T10:00:00Z"}
}'

# User 2: Jane Smith
aws dynamodb put-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--item '{
"entity_id": {"S": "USER-jane"},
"EntityType": {"S": "User"},
"Username": {"S": "jane"},
"Name": {"S": "Jane Smith"},
"Email": {"S": "jane@example.com"},
"PasswordHash": {"S": "$2a$10$examplehash"},
"CreatedAt": {"S": "2025-01-15T09:00:00Z"},
"UpdatedAt": {"S": "2025-01-15T09:00:00Z"}
}'

# User 3: Bob Wilson
aws dynamodb put-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--item '{
"entity_id": {"S": "USER-bob"},
"EntityType": {"S": "User"},
"Username": {"S": "bob"},
"Name": {"S": "Bob Wilson"},
"Email": {"S": "bob@example.com"},
"PasswordHash": {"S": "$2a$10$examplehash"},
"CreatedAt": {"S": "2025-02-01T08:00:00Z"},
"UpdatedAt": {"S": "2025-02-01T08:00:00Z"}
}'


# Create User Skills
# John's Skills
# Python - Expert, 7 years

aws dynamodb put-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--item '{
"entity_id": {"S": "SKILL-john-Python"},
"EntityType": {"S": "UserSkill"},
"Username": {"S": "john"},
"SkillName": {"S": "Python"},
"ProficiencyLevel": {"S": "Expert"},
"YearsOfExperience": {"N": "7"},
"Endorsements": {"N": "15"},
"LastUsedDate": {"S": "2025-12-10"},
"Notes": {"S": "Specialized in data science and ML"},
"CreatedAt": {"S": "2025-01-02T10:00:00Z"},
"UpdatedAt": {"S": "2025-12-10T10:00:00Z"}
}'

# JavaScript - Advanced, 5 years
aws dynamodb put-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--item '{
"entity_id": {"S": "SKILL-john-JavaScript"},
"EntityType": {"S": "UserSkill"},
"Username": {"S": "john"},
"SkillName": {"S": "JavaScript"},
"ProficiencyLevel": {"S": "Advanced"},
"YearsOfExperience": {"N": "5"},
"Endorsements": {"N": "8"},
"LastUsedDate": {"S": "2025-12-13"},
"Notes": {"S": "React and Node.js expert"},
"CreatedAt": {"S": "2025-01-02T10:00:00Z"},
"UpdatedAt": {"S": "2025-12-13T10:00:00Z"}
}'

# Jane's Skills
# Python - Expert, 10 years
aws dynamodb put-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--item '{
"entity_id": {"S": "SKILL-jane-Python"},
"EntityType": {"S": "UserSkill"},
"Username": {"S": "jane"},
"SkillName": {"S": "Python"},
"ProficiencyLevel": {"S": "Expert"},
"YearsOfExperience": {"S": "10"},
"Endorsements": {"N": "25"},
"LastUsedDate": {"S": "2025-12-12"},
"Notes": {"S": "Backend systems and APIs"},
"CreatedAt": {"S": "2025-01-16T09:00:00Z"},
"UpdatedAt": {"S": "2025-12-12T09:00:00Z"}
}'

# Go - Advanced, 6 years

aws dynamodb put-item \
--profile passbrains-ilisa-amplify \
--table-name glad-entities-production \
--item '{
"entity_id": {"S": "SKILL-jane-Go"},
"EntityType": {"S": "UserSkill"},
"Username": {"S": "jane"},
"SkillName": {"S": "Go"},
"ProficiencyLevel": {"S": "Advanced"},
"YearsOfExperience": {"S": "6"},
"Endorsements": {"N": "12"},
"LastUsedDate": {"S": "2025-12-11"},
"Notes": {"S": "Microservices and cloud infrastructure"},
"CreatedAt": {"S": "2025-01-16T09:00:00Z"},
"UpdatedAt": {"S": "2025-12-11T09:00:00Z"}
}'

# Bob's Skills
# Python - Intermediate, 3 years
aws dynamodb put-item \
--table-name glad-entities-production \
--item '{
"entity_id": {"S": "SKILL-bob-Python"},
"EntityType": {"S": "UserSkill"},
"Username": {"S": "bob"},
"SkillName": {"S": "Python"},
"ProficiencyLevel": {"S": "Intermediate"},
"YearsOfExperience": {"N": "3"},
"Endorsements": {"N": "5"},
"LastUsedDate": {"S": "2025-12-09"},
"Notes": {"S": "Learning Django framework"},
"CreatedAt": {"S": "2025-02-02T08:00:00Z"},
"UpdatedAt": {"S": "2025-12-09T08:00:00Z"}
}'

# JavaScript - Expert, 8 years
aws dynamodb put-item \
--table-name glad-entities-production \
--item '{
"entity_id": {"S": "SKILL-bob-JavaScript"},
"EntityType": {"S": "UserSkill"},
"Username": {"S": "bob"},
"SkillName": {"S": "JavaScript"},
"ProficiencyLevel": {"S": "Expert"},
"YearsOfExperience": {"N": "8"},
"Endorsements": {"N": "20"},
"LastUsedDate": {"S": "2025-12-13"},
"Notes": {"S": "Full-stack JavaScript developer"},
"CreatedAt": {"S": "2025-02-02T08:00:00Z"},
"UpdatedAt": {"S": "2025-12-13T08:00:00Z"}
}'

# TypeScript - Advanced, 4 years
aws dynamodb put-item \
--table-name glad-entities-production \
--item '{
"entity_id": {"S": "SKILL-bob-TypeScript"},
"EntityType": {"S": "UserSkill"},
"Username": {"S": "bob"},
"SkillName": {"S": "TypeScript"},
"ProficiencyLevel": {"S": "Advanced"},
"YearsOfExperience": {"N": "4"},
"Endorsements": {"N": "10"},
"LastUsedDate": {"S": "2025-12-13"},
"Notes": {"S": "Building enterprise applications"},
"CreatedAt": {"S": "2025-02-02T08:00:00Z"},
"UpdatedAt": {"S": "2025-12-13T08:00:00Z"}
}'

# Query Examples Using Composite Keys

# Find all Expert Python users
aws dynamodb query \
--profile passbrains-ilisa-amplify \
--table-name glad-entities-production \
--index-name SkillsByLevel \
--key-condition-expression "SkillName = :skill AND ProficiencyLevel = :level" \
--expression-attribute-values '{
":skill": {"S": "Python"},
":level": {"S": "Expert"}
}'

# Find Expert Python users with 5+ years experience
aws dynamodb query \
--table-name glad-entities-production \
--index-name SkillsByLevel \
--key-condition-expression "SkillName = :skill AND ProficiencyLevel = :level AND YearsOfExperience >= :years" \
--expression-attribute-values '{
":skill": {"S": "Python"},
":level": {"S": "Expert"},
":years": {"N": "5"}
}'

# Get all of John's profile and skills
aws dynamodb query \
--table-name glad-entities-production \
--index-name ByUser \
--key-condition-expression "Username = :user" \
--expression-attribute-values '{":user": {"S": "john"}}'

# Get only John's profile
aws dynamodb query \
--table-name glad-entities-production \
--index-name ByUser \
--key-condition-expression "Username = :user AND EntityType = :type" \
--expression-attribute-values '{
":user": {"S": "john"},
":type": {"S": "User"}
}'

# Get only John's skills
aws dynamodb query \
--table-name glad-entities-production \
--index-name ByUser \
--key-condition-expression "Username = :user AND EntityType = :type" \
--expression-attribute-values '{
":user": {"S": "john"},
":type": {"S": "UserSkill"}
}'

aws dynamodb put-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--item '{
"entity_id": {"S": "SKILL-Python"},
"EntityType": {"S": "Skill"},
"SkillName": {"S": "Python"},
"Description": {"S": "High-level programming language for general-purpose programming"},
"Category": {"S": "Programming"},
"CreatedAt": {"S": "2025-01-01T00:00:00Z"},
"UpdatedAt": {"S": "2025-01-01T00:00:00Z"}
}'

# Get all available skills (not user-specific)
aws dynamodb query \
--table-name glad-entities-production \
--index-name ByEntityType \
--profile passbrains-ilisa-amplify \
--key-condition-expression "EntityType = :type" \
--expression-attribute-values '{":type": {"S": "Skill"}}'

# Get a specific skill
aws dynamodb get-item \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--key '{"entity_id": {"S": "SKILL-Python"}}'

# Get all skills in Programming category
aws dynamodb query \
--table-name glad-entities-production \
--profile passbrains-ilisa-amplify \
--index-name SkillsByCategory \
--key-condition-expression "EntityType = :type AND Category = :cat" \
--expression-attribute-values '{
":type": {"S": "Skill"},
":cat": {"S": "Programming"}
}'
