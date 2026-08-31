import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers/core_providers.dart';
import 'location_models.dart';

class LocationRepository {
  LocationRepository(this._dio);
  final Dio _dio;

  Future<List<LocationModel>> listByActivity(String activityId) async {
    final response = await _dio.get('/locations', queryParameters: {'activity_id': activityId});
    return (response.data as List<dynamic>).map((e) => LocationModel.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<void> create({
    required String activityId,
    required String name,
    required double latitude,
    required double longitude,
    required int radiusMeters,
  }) async {
    await _dio.post('/locations', data: {
      'activity_id': activityId,
      'name': name,
      'latitude': latitude,
      'longitude': longitude,
      'radius_meters': radiusMeters,
    });
  }
}

final locationRepositoryProvider = Provider<LocationRepository>((ref) {
  return LocationRepository(ref.watch(apiClientProvider).dio);
});
