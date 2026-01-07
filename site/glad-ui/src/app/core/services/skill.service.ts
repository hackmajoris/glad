import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { HttpParams } from '@angular/common/http';
import { ApiService } from './api.service';
import { Skill, CreateSkillRequest, UpdateSkillRequest, UserSkillListItem, ProficiencyLevel } from '../models';

@Injectable({
  providedIn: 'root'
})
export class SkillService {
  constructor(private apiService: ApiService) {}

  /**
   * Add a new skill to a user
   */
  addSkill(username: string, request: CreateSkillRequest): Observable<Skill> {
    return this.apiService.post<Skill>(`/users/${username}/skills`, request);
  }

  /**
   * Get a specific skill for a user
   */
  getSkill(username: string, skillName: string): Observable<Skill> {
    return this.apiService.get<Skill>(`/users/${username}/skills/${skillName}`);
  }

  /**
   * List all skills for a user
   */
  listSkillsForUser(username: string): Observable<Skill[]> {
    return this.apiService.get<Skill[]>(`/users/${username}/skills`);
  }

  /**
   * Update an existing skill
   */
  updateSkill(username: string, skillName: string, request: UpdateSkillRequest): Observable<Skill> {
    return this.apiService.put<Skill>(`/users/${username}/skills/${skillName}`, request);
  }

  /**
   * Delete a skill
   */
  deleteSkill(username: string, skillName: string): Observable<{ message: string }> {
    return this.apiService.delete<{ message: string }>(`/users/${username}/skills/${skillName}`);
  }

  /**
   * Find users by skill name
   */
  listUsersBySkill(category: string, skillName: string, level?: ProficiencyLevel): Observable<UserSkillListItem[]> {
    let params = new HttpParams()
      .set('category', category);

    if (level) {
      params = params.set('level', level);
    }

    return this.apiService.get<UserSkillListItem[]>(`/skills/${skillName}/users`, params);
  }
}
