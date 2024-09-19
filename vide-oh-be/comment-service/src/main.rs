#![feature(decl_macro, proc_macro_hygiene)]
#[macro_use]
extern crate diesel;
extern crate dotenv;
#[macro_use]
extern crate rocket;
#[macro_use]
extern crate serde_derive;

use dotenv::dotenv;
use lambda_web::{is_running_on_lambda, launch_rocket_on_lambda, LambdaError};
use diesel_async::{AsyncPgConnection, AsyncConnection};
use diesel_async_migrations::EmbeddedMigrations;

use crate::connection::get_connection_string;

mod schema;
mod connection;
mod models;
mod repository;
mod handler;
mod router;
mod auth;
mod cors;

pub const MIGRATIONS: EmbeddedMigrations = diesel_async_migrations::embed_migrations!();

// Runs the database migrations
async fn run_migrations(connection_string: String) -> anyhow::Result<()> {
    let mut conn = AsyncPgConnection::establish(&connection_string).await?;
    MIGRATIONS.run_pending_migrations(&mut conn).await?;
    Ok(())
}

#[rocket::main]
async fn main() -> Result<(), LambdaError> {
    println!("Starting application...");

    dotenv().ok();

    // Fetch connection string from environment
    let connection_string = get_connection_string().await.expect("Failed to get DB connection string");

    // Run database migrations
    if let Err(e) = run_migrations(connection_string.clone()).await {
        eprintln!("Failed to run migrations: {:?}", e);
        return Err(LambdaError::from(e));
    }

    // Create and launch the Rocket application
    let rocket = router::create_routes(&connection_string).await?;
    println!("Before launching application");

    if is_running_on_lambda() {
        println!("Running on AWS Lambda");
        launch_rocket_on_lambda(rocket).await?;
    } else {
        println!("Running locally");
        rocket.launch().await?;
    }

    Ok(())
}
