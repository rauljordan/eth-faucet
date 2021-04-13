import React, { FormEvent, useState } from 'react';
import Avatar from '@material-ui/core/Avatar';
import Button from '@material-ui/core/Button';
import CssBaseline from '@material-ui/core/CssBaseline';
import TextField from '@material-ui/core/TextField';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/core/styles';
import Container from '@material-ui/core/Container';
import {LinearProgress, Snackbar} from '@material-ui/core';
import MuiAlert from '@material-ui/lab/Alert';

import { FundsRequest, FundsResponse, requestFunds } from '../services/requestFunds';
import { useRecaptcha } from '../hooks/recaptcha';
import { environment } from '../environment';

const useStyles = makeStyles((theme) => ({
    paper: {
        marginTop: theme.spacing(8),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    avatar: {
        margin: theme.spacing(1),
    },
    form: {
        width: '100%', // IE 11.
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
}));

const Alert = (props: any) => {
    return <MuiAlert elevation={6} variant="filled" {...props} />;
}

const FaucetForm = () => {
    const classes = useStyles();
    const [address, setAddress] = useState("");
    const [fundsResponse, setFundsResponse] = useState({} as FundsResponse);
    const [inProgress, setInProgress] = useState(false);
    const [openErr, setOpenErr] = React.useState(false);
    const [openFunded, setOpenFunded] = React.useState(false);
    const [error, setError] = React.useState(undefined);
    const executeCaptcha = useRecaptcha({
        sitekey: environment.captchaSiteKey,
    });

    const handleSubmit = (e: FormEvent) => doRequestFunds(e);
    const doRequestFunds = async (e: FormEvent) => {
        e.preventDefault();
        try {
            setInProgress(true);
            const token = await executeCaptcha(address);
            const res = await requestFunds({
                walletAddress: address,
                captchaResponse: token,
            } as FundsRequest)
            setFundsResponse(res);
            setOpenFunded(true);
            setInProgress(false);
        } catch (err) {
            if (err.response && err.response.data && err.response.data.message) {
                setOpenErr(true);
                setError(err.response.data.message);
            } else {
                setError(err);
            }
            setInProgress(false);
        }
    };
    return (
        <>
            { inProgress && <LinearProgress color="secondary" /> }
            <Container component="main" maxWidth="xs">
                <CssBaseline />
                <div className={classes.paper}>
                    <Avatar
                        className={classes.avatar}
                        src="https://ih1.redbubble.net/image.529445044.0787/st,small,845x845-pad,1000x1000,f8f8f8.u7.jpg">
                    </Avatar>
                    <Typography component="h1" variant="h5">
                        {
                            inProgress ? 'Requesting...' : 'Request Faucet Funds'
                        }
                    </Typography>
                    {
                        fundsResponse.transactionHash &&
                        <div>
                            <a href={`https://goerli.etherscan.io/tx/${fundsResponse.transactionHash}`}>
                                View Transaction on Etherscan
                            </a>
                        </div>
                    }
                    <form className={classes.form} noValidate onSubmit={handleSubmit}>
                        <TextField
                            variant="outlined"
                            margin="normal"
                            required
                            fullWidth
                            id="address"
                            label="ETH Address"
                            name="address"
                            value={address}
                            onChange={(e) => setAddress(e.target.value)}
                            autoFocus
                        />
                        <Button
                            type="submit"
                            fullWidth
                            variant="contained"
                            color="primary"
                            disabled={inProgress}
                            className={classes.submit}
                        >
                            Request ETH
                        </Button>
                    </form>
                </div>
            </Container>
            <Snackbar open={openFunded} autoHideDuration={3000}>
                <Alert severity="success">
                    Funded with {fundsResponse.amount}
                </Alert>
            </Snackbar>
            <Snackbar open={openErr} autoHideDuration={3000}>
                <Alert severity="error">
                    ERROR: {error}
                </Alert>
            </Snackbar>
        </>
    );
};

export default FaucetForm;