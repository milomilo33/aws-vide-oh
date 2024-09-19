use rocket::http::Status;
use rocket::serde::json::Json;
use crate::auth::MyJWTClaims;
use crate::connection::DbConn;
use crate::models::Comment;
use crate::models::NewComment;
use crate::models::NewRating;
use crate::repository;


#[get("/comments/<video_id>")]
pub async fn show_all_comments_for_video<'a>(video_id: i32, mut connection: DbConn<'a>, _my_claims: MyJWTClaims) -> Result<Json<Vec<Comment>>, Status> {
    match repository::show_all_comments_for_video(video_id, &mut connection).await {
        Ok(comments) => Ok(Json(comments)),
        Err(_) => Err(Status::NotFound),
    }
}

#[post("/comments", format = "application/json", data = "<new_comment>")]
pub async fn create_comment<'a>(new_comment: Json<NewComment>, mut connection: DbConn<'a>, my_claims: MyJWTClaims) -> Result<Status, Status> {
    println!("Got into posting comment1");
    if !my_claims.email.eq(&new_comment.owner_email) {
        return Err(Status::Unauthorized);
    }
    println!("Got into posting comment2");
    match repository::create_comment(new_comment.into_inner(), &mut connection).await {
        Ok(_) => Ok(Status::Ok),
        Err(err) => {
            // Print the error message
            println!("Failed to create comment: {:?}", err);
            Err(Status::BadRequest)
        },
    }
}

#[get("/comments/reported")]
pub async fn show_all_reported_comments<'a>(mut connection: DbConn<'a>, my_claims: MyJWTClaims) -> Result<Json<Vec<Comment>>, Status> {
    if !my_claims.role.eq(&String::from("Administrator")) {
        return Err(Status::Unauthorized);
    }

    match repository::show_all_reported_comments(&mut connection).await {
        Ok(comments) => Ok(Json(comments)),
        Err(_) => Err(Status::NotFound),
    }
}

#[get("/comments/delete/<comment_id>")]
pub async fn delete_comment<'a>(comment_id: i32, mut connection: DbConn<'a>, my_claims: MyJWTClaims) -> Result<Status, Status> {
    match repository::get_comment(comment_id, &mut connection).await {
        Ok(comment) => {
            if my_claims.role.eq(&String::from("RegisteredUser")) && !my_claims.email.eq(&comment.owner_email) {
                return Err(Status::Unauthorized);
            }

            match repository::delete_comment(comment_id, &mut connection).await {
                Ok(_) => Ok(Status::Ok),
                Err(_) => Err(Status::NotFound),
            }
        },
        Err(_) => Err(Status::NotFound),
    }
}

#[get("/comments/report/<comment_id>")]
pub async fn report_comment<'a>(comment_id: i32, mut connection: DbConn<'a>, _my_claims: MyJWTClaims) -> Result<Status, Status> {
    match repository::report_comment(comment_id, &mut connection).await {
        Ok(_) => Ok(Status::Ok),
        Err(_) => Err(Status::NotFound),
    }
}

#[post("/ratings", format = "application/json", data = "<new_rating>")]
pub async fn create_or_update_rating<'a>(new_rating: Json<NewRating>, mut connection: DbConn<'a>, my_claims: MyJWTClaims) -> Result<Status, Status> {
    if !my_claims.email.eq(&new_rating.rating_owner_email) {
        return Err(Status::Unauthorized);
    }

    match repository::create_or_update_rating(new_rating.into_inner(), &mut connection).await {
        Ok(_) => Ok(Status::Ok),
        Err(_) => Err(Status::BadRequest),
    }
}

#[get("/ratings/total/<video_id>")]
pub async fn get_rating_for_video<'a>(video_id: i32, mut connection: DbConn<'a>, _my_claims: MyJWTClaims) -> Result<Json<f32>, Status> {
    match repository::get_rating_for_video(video_id, &mut connection).await {
        Ok(rat) => Ok(Json(rat)),
        Err(_) => Err(Status::NotFound),
    }
}

#[get("/ratings/user/<owner_email>/<video_id>")]
pub async fn get_rating_for_user<'a>(owner_email: String, video_id: i32, mut connection: DbConn<'a>, my_claims: MyJWTClaims) -> Result<Json<i32>, Status> {
    if !my_claims.email.eq(&owner_email) {
        return Err(Status::Unauthorized);
    }

    match repository::get_rating_for_user(owner_email, video_id, &mut connection).await {
        Ok(rat) => Ok(Json(rat)),
        Err(_) => Err(Status::NotFound),
    }
}
