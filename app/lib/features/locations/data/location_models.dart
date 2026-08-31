class LocationModel {
  LocationModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        activityId = json['activity_id'] as String,
        name = json['name'] as String,
        latitude = (json['latitude'] as num).toDouble(),
        longitude = (json['longitude'] as num).toDouble(),
        radiusMeters = json['radius_meters'] as int;

  final String id;
  final String activityId;
  final String name;
  final double latitude;
  final double longitude;
  final int radiusMeters;
}
