class ActivityModel {
  ActivityModel.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        name = json['name'] as String,
        description = json['description'] as String? ?? '',
        isActive = json['is_active'] as bool;

  final String id;
  final String name;
  final String description;
  final bool isActive;
}
