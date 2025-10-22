CREATE TABLE IF NOT EXISTS event (
	id 			 INTEGER PRIMARY KEY,
	start_time 	 INTEGER 	  NOT NULL,
	window_class VARCHAR(255) NOT NULL,
	window_title VARCHAR(255),
	duration 	 INTEGER      NOT NULL
);
