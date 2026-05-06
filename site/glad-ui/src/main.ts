import { bootstrapApplication } from '@angular/platform-browser';
import { Amplify } from 'aws-amplify';
import { appConfig } from './app/app.config';
import { App } from './app/app';
import amplifyOutputs from '../amplify_outputs.json';

// Configure Amplify
Amplify.configure(amplifyOutputs);

bootstrapApplication(App, appConfig)
  .catch((err) => console.error(err));
