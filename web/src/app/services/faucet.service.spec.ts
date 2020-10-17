import { TestBed } from '@angular/core/testing';

import { FaucetService } from './faucet.service';

describe('FaucetService', () => {
  let service: FaucetService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(FaucetService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
