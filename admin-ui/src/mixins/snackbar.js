const snackbar = {
	data: () => ({
		snackbar: {
			visibility: false,
			message: '',
			timeout: 3000,
			route: {},
			color: '',
			buttonColor: 'primary'
		}
	}),
	methods: {
		showSnackbar({message, timeout = 3000, route = {}, color = '', buttonColor = 'primary'}){
			this.snackbar.message = message;
			this.snackbar.timeout = timeout;
			this.snackbar.route = route;
			this.snackbar.visibility = true;
			this.snackbar.color = color;
			this.snackbar.buttonColor = buttonColor;
		},
		hideSnackbar(){
			this.snackbar.visibility = false
		},
		showSnackbarError({message, timeout}){
			const opts = {
				message,
				timeout,
				color: 'red darken-3',
				buttonColor: 'white'
			}
			this.showSnackbar(opts);
		},
		showSnackbarSuccess({message, timeout}){
			const opts = {
				message,
				timeout,
				color: 'green darken-3',
				buttonColor: 'white'
			}
			this.showSnackbar(opts);
		}
	}
}


export default snackbar;