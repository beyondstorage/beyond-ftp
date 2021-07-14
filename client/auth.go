package client

// Handle the "USER" command.
func (c *Handler) handleUSER() {
	c.user = c.param
	c.WriteMessage(StatusUserOK, "User name okay, need password.")
}

// Handle the "PASS" command.
func (c *Handler) handlePASS() {
	if c.user == "" {
		c.WriteMessage(StatusBadCommandSequence, "User is expected before Pass")
		return
	}

	defer func() {
		c.user = ""
	}()

	username := c.user
	password := c.param

	if _, ok := c.serverSetting.Users[username]; ok {
		if username == "anonymous" || password == c.serverSetting.Users[username] {
			c.loginUser = username
			c.WriteMessage(StatusUserLoggedIn, "Password ok, continue")
			return
		}
	}

	c.WriteMessage(StatusNotLoggedIn, "Invalid username or password")
}
