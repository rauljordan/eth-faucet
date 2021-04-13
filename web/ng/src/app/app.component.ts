import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { throwError } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';
import { FaucetService, FundsResponse } from './services/faucet.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  loading = false;
  formGroup: FormGroup;
  transactionHash: string;
  constructor(
    private readonly faucetService: FaucetService,
    private readonly snackbar: MatSnackBar,
    private readonly fb: FormBuilder,
  ) { }
  ngOnInit(): void {
    this.formGroup = this.fb.group({
      walletAddress: new FormControl('', [
        Validators.required,
      ]),
    });
  }
  submit(e: Event): void {
    e.stopPropagation();
    this.formGroup.markAsDirty();
    if (this.formGroup.invalid) {
      return;
    }
    this.loading = true;
    const addr = this.formGroup.controls.walletAddress.value;
    this.faucetService.requestFunds(addr).pipe(
      tap((res: FundsResponse) => {
        this.loading = false;
        this.transactionHash = res.transactionHash;
        const message = `Funded with ${res.amount} ETH`;
        this.snackbar.open(message, 'Close', {
          duration: 2000,
        });
      }),
      catchError((err: HttpErrorResponse) => {
        this.loading = false;
        this.snackbar.open(err.error.message, 'Close', {
          duration: 2000,
        });
        return throwError(err);
      }),
    ).subscribe();
  }
}
