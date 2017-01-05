/**
  * @author Andreas Schick (2792119), Linda Latreider (7743782), Niklas Nikisch (9364290)
  */

window.addEventListener("DOMContentLoaded", function(){
	//document.getElementById("loginButton").onclick(onClickLoginProcess);
	//document.getElementById("registerButton").onclick(onClickRegisterProcess);

	function getUrlParameter(paramName){
		var result = "-1",
			tmp = [];
		location
			.search.substr(1)
			.split("&")
			.forEach(function (item) {
				tmp = item.split("=");
				if (tmp[0] === paramName){
					result = decodeURIComponent(tmp[1]);
				}
		});
		return result;
	}
	
	window.onload=function validate(){
		var message = "none";
		if(getUrlParameter("register")==="userfalse"){
			var message="Registrierung fehlgeschlagen. Benutzername bereits vergeben.";
			document.getElementById("r_username").style.backgroundColor="lightpink";
		}
		if(getUrlParameter("register")==="pwfalse"){
			var message="Registrierung fehlgeschlagen. Passwortwiederholung nicht korrekt.";
			document.getElementById("r_password").style.backgroundColor="lightpink";
			document.getElementById("r_password2").style.backgroundColor="lightpink";
		}
		if(getUrlParameter("login")==="false"){
			var message="Anmeldung fehlgeschlagen. Nutzername oder Passwort nicht korrekt.";
			document.getElementById("l_username").style.backgroundColor="lightpink";
			document.getElementById("l_password").style.backgroundColor="lightpink";
		}
		
		var messagefield = document.getElementById("errormessage");
		if(message!=="none"){
			messagefield.style.display='style';
			messagefield.innerHTML = message;
		} else {
			messagefield.style.display='none';
			document.getElementById(errormessage).setActive(false);
		}
	}
})