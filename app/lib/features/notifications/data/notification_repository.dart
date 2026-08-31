import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/providers/core_providers.dart';

class NotificationModel {
  NotificationModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        title = json['title'] as String,
        body = json['body'] as String,
        type = json['type'] as String,
        isRead = json['is_read'] as bool,
        // .toLocal() matters here: this is a real point-in-time timestamp
        // (unlike a pure calendar date), so it must render in the viewer's
        // timezone or an early-morning IST notification displays as
        // "yesterday" once DateFormat renders the UTC-flagged instant.
        createdAt = DateTime.parse(json['created_at'] as String).toLocal();

  final String id;
  final String title;
  final String body;
  final String type;
  final bool isRead;
  final DateTime createdAt;
}

class NotificationRepository {
  NotificationRepository(this._dio);
  final Dio _dio;

  Future<PagedResult<NotificationModel>> list({int page = 1}) async {
    final response = await _dio.get('/notifications', queryParameters: {'page': page});
    return PagedResult.fromJson(response.data as Map<String, dynamic>, NotificationModel.fromJson);
  }

  Future<void> markRead(String id) => _dio.post('/notifications/$id/read');
}

final notificationRepositoryProvider = Provider<NotificationRepository>((ref) {
  return NotificationRepository(ref.watch(apiClientProvider).dio);
});

final notificationsListProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(notificationRepositoryProvider).list();
});
