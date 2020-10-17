import { InjectionToken } from '@angular/core';

export interface IEnvironment {
    production: boolean;
    apiEndpoint: string;
    catpchaSiteKey: string;
}

export const ENVIRONMENT = new InjectionToken<IEnvironment>('ENVIRONMENT');