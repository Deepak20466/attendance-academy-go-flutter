import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../activities/data/activity_repository.dart';
import '../data/location_repository.dart';

class LocationsScreen extends ConsumerStatefulWidget {
  const LocationsScreen({super.key});

  @override
  ConsumerState<LocationsScreen> createState() => _LocationsScreenState();
}

class _LocationsScreenState extends ConsumerState<LocationsScreen> {
  String? _activityId;

  @override
  Widget build(BuildContext context) {
    final activities = ref.watch(activitiesListProvider);
    // Resolved up front (not mutated as a side effect deep in `body`) so the
    // FAB — built before `body` in this Scaffold — sees the same value the
    // dropdown defaults to on the very first frame, instead of one build late.
    final loadedActivities = activities.valueOrNull;
    final resolvedActivityId = _activityId ?? (loadedActivities != null && loadedActivities.isNotEmpty ? loadedActivities.first.id : null);

    return Scaffold(
      appBar: AppBar(title: const Text('Locations')),
      floatingActionButton: resolvedActivityId == null
          ? null
          : FloatingActionButton.extended(
              onPressed: () => _showCreateDialog(context, ref, resolvedActivityId),
              icon: const Icon(Icons.add),
              label: const Text('Add location'),
            ),
      body: AsyncValueWidget(
        value: activities,
        data: (items) {
          if (items.isEmpty) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(32),
                child: Text('Create an activity first before adding locations.', style: TextStyle(color: Colors.grey)),
              ),
            );
          }
          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.all(16),
                child: DropdownButtonFormField<String>(
                  initialValue: resolvedActivityId,
                  decoration: const InputDecoration(labelText: 'Activity'),
                  items: items.map((a) => DropdownMenuItem(value: a.id, child: Text(a.name))).toList(),
                  onChanged: (v) => setState(() => _activityId = v),
                ),
              ),
              Expanded(child: _LocationsList(activityId: resolvedActivityId!)),
            ],
          );
        },
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref, String activityId) async {
    final nameController = TextEditingController();
    final latController = TextEditingController();
    final lngController = TextEditingController();
    final radiusController = TextEditingController(text: '100');
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Add location'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(controller: nameController, decoration: const InputDecoration(labelText: 'Name')),
                const SizedBox(height: 12),
                TextField(
                  controller: latController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true, signed: true),
                  decoration: const InputDecoration(labelText: 'Latitude'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: lngController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true, signed: true),
                  decoration: const InputDecoration(labelText: 'Longitude'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: radiusController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(labelText: 'Allowed radius (meters)'),
                ),
                if (errorText != null) ...[
                  const SizedBox(height: 12),
                  Text(errorText!, style: const TextStyle(color: Colors.red)),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            FilledButton(
              onPressed: () async {
                final lat = double.tryParse(latController.text);
                final lng = double.tryParse(lngController.text);
                final radius = int.tryParse(radiusController.text);
                if (nameController.text.trim().isEmpty || lat == null || lng == null || radius == null || radius <= 0) {
                  setState(() => errorText = 'Enter a valid name, latitude, longitude and radius.');
                  return;
                }
                try {
                  await ref.read(locationRepositoryProvider).create(
                        activityId: activityId,
                        name: nameController.text.trim(),
                        latitude: lat,
                        longitude: lng,
                        radiusMeters: radius,
                      );
                  ref.invalidate(_locationsForActivityProvider(activityId));
                  if (context.mounted) Navigator.pop(context);
                } catch (e) {
                  setState(() => errorText = 'Failed: $e');
                }
              },
              child: const Text('Add'),
            ),
          ],
        ),
      ),
    );
  }
}

class _LocationsList extends ConsumerWidget {
  const _LocationsList({required this.activityId});
  final String activityId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final locations = ref.watch(_locationsForActivityProvider(activityId));
    return RefreshIndicator(
      onRefresh: () => ref.refresh(_locationsForActivityProvider(activityId).future),
      child: AsyncValueWidget(
        value: locations,
        data: (items) {
          if (items.isEmpty) {
            return ListView(
              children: const [
                Padding(
                  padding: EdgeInsets.all(32),
                  child: Center(child: Text('No locations found for this activity.', style: TextStyle(color: Colors.grey))),
                ),
              ],
            );
          }
          return ListView.separated(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: items.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final l = items[index];
              return ListTile(
                leading: const Icon(Icons.location_on_outlined),
                title: Text(l.name),
                subtitle: Text('${l.latitude.toStringAsFixed(5)}, ${l.longitude.toStringAsFixed(5)} • ${l.radiusMeters}m radius'),
              );
            },
          );
        },
      ),
    );
  }
}

final _locationsForActivityProvider = FutureProvider.autoDispose.family((ref, String activityId) {
  return ref.watch(locationRepositoryProvider).listByActivity(activityId);
});
