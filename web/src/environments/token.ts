import { InjectionToken } from '@angular/core';

export interface IEnvironment {
    production: boolean;
}

export const ENVIRONMENT = new InjectionToken<IEnvironment>('ENVIRONMENT');