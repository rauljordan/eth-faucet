import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { ENVIRONMENT } from '../environments/token';
import { environment } from '../environments/environment';
import { AppComponent } from './app.component';

@NgModule({
  declarations: [
    AppComponent
  ],
  imports: [
    BrowserModule,
  ],
  providers: [
    { provide: ENVIRONMENT, useValue: environment },
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
