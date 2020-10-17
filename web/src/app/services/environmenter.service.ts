import { Inject, Injectable } from '@angular/core';
import { ENVIRONMENT, IEnvironment } from 'src/environments/token';

@Injectable({
  providedIn: 'root'
})
export class EnvironmenterService {
  constructor(
    @Inject(ENVIRONMENT) private environment: IEnvironment,
  ) {}
  public readonly env = this.environment;
}