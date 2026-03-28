<H1> Sample Code:</H1>

//Entry point of the web app:
App {
    LoginScreen()
}


//Starting Screen
fun LoginScreen() {
    set pageTitle = "DriftUI | Secure Authentication"
    
    val nav = useNav()
    @State var email = ""
    @State var password = ""

    // Root Container - Neutral Background
    Box(modifier = Modifier.fillMaxSize().background("#F9FAFB")) {
        
        // Centered Auth Card
        Column(
            modifier = Modifier
                .align("center")
                .width(420.dp)
                .background(Color.White)
                .cornerRadius(16.dp)
                .shadow(24.dp)
                .padding(40.dp, 40.dp, 40.dp, 40.dp)
        ) {
            // Branding/Logo Placeholder
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .background("#111827")
                    .cornerRadius(12.dp)
            ) {}
            
            Text(
                "Welcome back", 
                modifier = Modifier.fontSize(28).fontWeight("800").color("#111827").padding(24.dp, 0, 8.dp, 0)
            )
            Text(
                "Enter your credentials to access your workspace.", 
                modifier = Modifier.fontSize(15).color(Color.gray).padding(0, 0, 32.dp, 0)
            )

            // Email Input
            Text("Email", modifier = Modifier.fontSize(14).fontWeight("600").color("#374151").padding(0, 0, 8.dp, 0))
            TextField(
                value = email,
                onValueChange = { email = it },
                singleLine = true,
                placeholder = { Text("name@company.com") },
                modifier = Modifier.fillMaxWidth().padding(0, 0, 20.dp, 0)
            )

            // Password Input
            Text("Password", modifier = Modifier.fontSize(14).fontWeight("600").color("#374151").padding(0, 0, 8.dp, 0))
            TextField(
                value = password,
                onValueChange = { password = it },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier.fillMaxWidth().padding(0, 0, 24.dp, 0)
            )

            // Utility Row
            Row(modifier = Modifier.fillMaxWidth().padding(0, 0, 32.dp, 0)) {
                Box(modifier = Modifier.weight(1.f)) {} // Spacer
                Text(
                    "Forgot password?", 
                    modifier = Modifier.fontSize(14).fontWeight("600").color("#3B82F6").clickable {
                        if (email == "") {
                            Toast(
                                message = "Enter your email address first.", 
                                duration = ToastLength.Short, 
                                type = ToastType.Error
                            )
                        } else {
                            Toast(
                                message = "Recovery link sent to " + email, 
                                duration = ToastLength.Long, 
                                type = ToastType.Info
                            )
                        }
                    }
                )
            }

            // Primary Action
            Button(
                onClick = {
                    if (email == "" || password == "") {
                        Toast(
                            message = "Please complete all fields.", 
                            duration = ToastLength.Short, 
                            type = ToastType.Error
                        )
                    } else {
                        Toast(
                            message = "Authenticating...", 
                            duration = ToastLength.Short, 
                            type = ToastType.Success
                        )
                        nav.replaceRoot(tag = "Dashboard") { DashboardScreen() }
                    }
                },
                modifier = Modifier.fillMaxWidth().background("#111827").padding(0, 0, 24.dp, 0)
            ) {
                Text("Sign In", modifier = Modifier.color(Color.White).fontWeight("600").fontSize(15))
            }
            
            // Footer Navigation
            Row(modifier = Modifier.fillMaxWidth().align("center")) {
                Text("Don't have an account? ", modifier = Modifier.fontSize(14).color(Color.gray))
                Text(
                    "Sign up", 
                    modifier = Modifier.fontSize(14).fontWeight("600").color("#111827").clickable {
                        Toast(message = "Redirecting to Sign Up...", type = ToastType.Info)
                    }
                )
            }
        }
    }
}

fun DashboardScreen() {
    set pageTitle = "Dashboard"
    
    val nav = useNav()
    
    Column(modifier = Modifier.fillMaxSize().background(Color.White).padding(48.dp, 48.dp, 48.dp, 48.dp)) {
        Text("Workspace Initialized", modifier = Modifier.fontSize(32).fontWeight("800").padding(0, 0, 32.dp, 0))
        
        Button(
            onClick = { nav.replaceRoot(tag = "Login") { LoginScreen() } },
            modifier = Modifier.background(Color.Red)
        ) {
            Text("Log Out", modifier = Modifier.color(Color.White).fontWeight("600"))
        }
    }
}
