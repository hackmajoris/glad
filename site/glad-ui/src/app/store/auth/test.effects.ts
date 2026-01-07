import { Injectable } from '@angular/core';
import { Actions, createEffect } from '@ngrx/effects';
import { EMPTY } from 'rxjs';

@Injectable()
export class TestEffects {
  constructor(private actions$: Actions) {
    console.log('TestEffects constructor, actions$:', this.actions$);
  }

  test$ = createEffect(() => this.actions$.pipe(), { dispatch: false });
}