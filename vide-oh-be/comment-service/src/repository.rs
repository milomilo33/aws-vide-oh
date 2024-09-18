use diesel_async::RunQueryDsl;
use diesel_async::AsyncPgConnection;
use crate::models::{Comment, NewComment, Rating, NewRating};
use crate::schema::comments;
use crate::schema::comments::dsl::*;
use crate::schema::ratings;
use crate::schema::ratings::dsl::*;
use diesel::prelude::*;
use diesel::result::QueryResult;

pub async fn get_comment(comment_id: i32, connection: &mut AsyncPgConnection) -> QueryResult<Comment> {
    comments::table.find(comment_id)
        .first(connection).await
}

pub async fn create_comment(new_comment: NewComment, conn: &mut AsyncPgConnection) -> QueryResult<Comment> {
    diesel::insert_into(comments::table)
        .values(&new_comment)
        .get_result(conn).await
}

pub async fn show_all_comments_for_video(video_id_provided: i32, connection: &mut AsyncPgConnection) -> QueryResult<Vec<Comment>> {
    comments.filter(video_id.eq(video_id_provided))
        .load::<Comment>(connection).await
}

pub async fn show_all_reported_comments(connection: &mut AsyncPgConnection) -> QueryResult<Vec<Comment>> {
    comments.filter(reported.eq(true))
        .load::<Comment>(connection).await
}

pub async fn delete_comment(comment_id: i32, connection: &mut AsyncPgConnection) -> QueryResult<usize> {
    diesel::delete(comments::table.find(comment_id))
        .execute(connection).await
}

pub async fn report_comment(comment_id: i32, connection: &mut AsyncPgConnection) -> QueryResult<Comment> {
    diesel::update(comments::table.find(comment_id))
        .set(reported.eq(true))
        .get_result(connection).await
}

pub async fn create_or_update_rating(new_rating: NewRating, conn: &mut AsyncPgConnection) -> QueryResult<Rating> {
    match ratings.filter(rating_owner_email.eq(&new_rating.rating_owner_email))
        .filter(rating_video_id.eq(&new_rating.rating_video_id))
        .select(rating_id)
        .first::<i32>(conn).await {
            Ok(found_rating_id) => {
                diesel::update(ratings::table.find(found_rating_id))
                .set(rating.eq(new_rating.rating))
                .get_result(conn).await
            },
            Err(_) => {
                diesel::insert_into(ratings::table)
                .values(&new_rating)
                .get_result(conn).await
            }
        }
}

pub async fn get_rating_for_video(video_id_provided: i32, connection: &mut AsyncPgConnection) -> QueryResult<f32> {
    match ratings.filter(rating_video_id.eq(video_id_provided))
        .select(rating)
        .load::<i32>(connection).await {
            Ok(ratings_vec) => {
                if ratings_vec.len() == 0 {
                    return Ok(0.0);
                }

                let rating_sum: i32 = ratings_vec.iter().sum();
                Ok(rating_sum as f32 / ratings_vec.len() as f32)
            },
            Err(err) => Err(err)
        }
}

pub async fn get_rating_for_user(user_email_provided: String, video_id_provided: i32, connection: &mut AsyncPgConnection) -> QueryResult<i32> {
    match ratings.filter(rating_owner_email.eq(user_email_provided))
        .filter(rating_video_id.eq(&video_id_provided))
        .select(rating)
        .first::<i32>(connection).await {
            Ok(found_rating) => Ok(found_rating),
            Err(err) => Err(err)
        }
}
