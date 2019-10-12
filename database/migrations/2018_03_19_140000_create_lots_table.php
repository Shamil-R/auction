<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateLotsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('lots', function (Blueprint $table) {
            $table->increments('id');
            $table->text('rules');
            $table->dateTime('created_at');
            $table->dateTime('updated_at');
            $table->dateTime('deleted_at')->nullable();
            $table->dateTime('booked_at')->nullable();
            $table->boolean('manual_booked')->default(false);
            $table->dateTime('confirmed_at')->nullable();
            $table->dateTime('completed_at')->nullable();
            $table->unsignedInteger('object_id');
            $table->text('object');
            $table->text('confirm')->nullable();
            $table->text('complete')->nullable();
            $table->string('group_key', 10);
            $table->unsignedInteger('user_id');

            $table->unique(['group_key', 'object_id']);

            $table->index('group_key');
            $table->index('user_id');

            $table->foreign('group_key')->references('group_key')->on('groups');
            $table->foreign('user_id')->references('id')->on('users');
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('lots');
    }
}
