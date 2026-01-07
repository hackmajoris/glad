export type ProficiencyLevel = 'Beginner' | 'Intermediate' | 'Advanced' | 'Expert';

export interface Skill {
  skillName: string;
  proficiencyLevel: ProficiencyLevel;
  yearsOfExperience: number;
  endorsements: number;
  lastUsedDate: string;
  notes: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateSkillRequest {
  skillName: string;
  proficiencyLevel: ProficiencyLevel;
  yearsOfExperience: number;
  notes?: string;
}

export interface UpdateSkillRequest {
  proficiencyLevel?: ProficiencyLevel;
  yearsOfExperience?: number;
  notes?: string;
}

export interface UserSkillListItem {
  username: string;
  name: string;
  proficiencyLevel: ProficiencyLevel;
}
