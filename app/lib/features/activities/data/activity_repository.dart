import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers/core_providers.dart';
import 'activity_models.dart';

class ActivityRepository {
  ActivityRepository(this._dio);
  final Dio _dio;

  Future<List<ActivityModel>> list({bool onlyActive = true}) async {
    final response = await _dio.get('/activities', queryParameters: {'only_active': onlyActive});
    return (response.data as List<dynamic>).map((e) => ActivityModel.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<void> create({required String name, required String description}) async {
    await _dio.post('/activities', data: {'name': name, 'description': description, 'is_active': true});
  }
}

final activityRepositoryProvider = Provider<ActivityRepository>((ref) {
  return ActivityRepository(ref.watch(apiClientProvider).dio);
});

final activitiesListProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(activityRepositoryProvider).list();
});
